// Copyright 2017-2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cilium/cilium/pkg/completion"
	"github.com/cilium/cilium/pkg/flowdebug"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/kafka"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/policy"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/cilium/cilium/pkg/proxy/accesslog"
	"github.com/cilium/cilium/pkg/proxy/logger"

	"github.com/optiopay/kafka/proto"
	"github.com/sirupsen/logrus"
)

const (
	fieldID = "id"
)

// kafkaRedirect implements the Redirect interface for an l7 proxy
type kafkaRedirect struct {
	redirect *Redirect
	conf     kafkaConfiguration
	rules    policy.L7DataMap
	socket   *proxySocket
}

type destLookupFunc func(remoteAddr string, dport uint16) (uint32, string, error)

type kafkaConfiguration struct {
	noMarker      bool
	lookupNewDest destLookupFunc
}

// createKafkaRedirect creates a redirect to the kafka proxy. The redirect structure passed
// in is safe to access for reading and writing.
func createKafkaRedirect(r *Redirect, conf kafkaConfiguration) (RedirectImplementation, error) {
	redir := &kafkaRedirect{
		redirect: r,
		conf:     conf,
	}

	if redir.conf.lookupNewDest == nil {
		redir.conf.lookupNewDest = lookupNewDest
	}

	marker := 0
	if !conf.noMarker {
		markIdentity := int(0)
		// As ingress proxy, all replies to incoming requests must have the
		// identity of the endpoint we are proxying for
		if r.ingress {
			markIdentity = int(r.source.GetIdentity())
		}

		marker = GetMagicMark(r.ingress, markIdentity)
	}

	// Listen needs to be in the synchronous part of this function to ensure that
	// the proxy port is never refusing connections.
	socket, err := listenSocket(fmt.Sprintf(":%d", r.ProxyPort), marker)
	if err != nil {
		return nil, err
	}

	redir.socket = socket

	go func() {
		for {
			pair, err := socket.Accept(true)
			select {
			case <-socket.closing:
				// Don't report errors while the socket is being closed
				return
			default:
			}

			if err != nil {
				log.WithField(logfields.Port, r.ProxyPort).WithError(err).Error("Unable to accept connection on port")
				continue
			}

			go redir.handleRequestConnection(pair)
		}
	}()

	return redir, nil
}

// canAccess determines if the kafka message req sent by identity is allowed to
// be forwarded according to the rules configured on kafkaRedirect
func (k *kafkaRedirect) canAccess(req *kafka.RequestMessage, srcIdentity identity.NumericIdentity) bool {
	var id *identity.Identity

	if srcIdentity != 0 {
		id = identity.LookupIdentityByID(srcIdentity)
		if id == nil {
			log.WithFields(logrus.Fields{
				logfields.Request:  req.String(),
				logfields.Identity: srcIdentity,
			}).Warn("Unable to resolve identity to labels")
		}
	}

	scopedLog := log.WithFields(logrus.Fields{
		logfields.Request:  req.String(),
		logfields.Identity: id,
	})

	k.redirect.mutex.RLock()
	rules := k.redirect.rules.GetRelevantRules(id)
	k.redirect.mutex.RUnlock()

	if rules.Kafka == nil {
		flowdebug.Log(scopedLog, "No Kafka rules matching identity, rejecting")
		return false
	}

	b, err := json.Marshal(rules.Kafka)
	if err != nil {
		flowdebug.Log(scopedLog, "Error marshalling kafka rules to apply")
		return false
	} else {
		flowdebug.Log(scopedLog.WithField("rule", string(b)), "Applying rule")
	}

	return req.MatchesRule(rules.Kafka)
}

// kafkaLogRecord wraps an accesslog.LogRecord so that we can define methods with a receiver
type kafkaLogRecord struct {
	logger.LogRecord
	topics []string
}

func apiKeyToString(apiKey int16) string {
	if key, ok := api.KafkaReverseAPIKeyMap[apiKey]; ok {
		return key
	}
	return fmt.Sprintf("%d", apiKey)
}

func (k *kafkaRedirect) newLogRecordFromRequest(req *kafka.RequestMessage) kafkaLogRecord {
	return kafkaLogRecord{
		LogRecord: logger.NewLogRecord(k.redirect, accesslog.TypeRequest,
			logger.LogTags.Kafka(&accesslog.LogRecordKafka{
				APIVersion:    req.GetVersion(),
				APIKey:        apiKeyToString(req.GetAPIKey()),
				CorrelationID: int32(req.GetCorrelationID()),
			})),
		topics: req.GetTopics(),
	}
}

func (k *kafkaRedirect) newLogRecordFromResponse(res *kafka.ResponseMessage, req *kafka.RequestMessage) kafkaLogRecord {
	lr := kafkaLogRecord{
		LogRecord: logger.NewLogRecord(k.redirect, accesslog.TypeResponse,
			logger.LogTags.Kafka(&accesslog.LogRecordKafka{})),
	}

	if res != nil {
		lr.Kafka.CorrelationID = int32(res.GetCorrelationID())
	}

	if req != nil {
		lr.Kafka.APIVersion = req.GetVersion()
		lr.Kafka.APIKey = apiKeyToString(req.GetAPIKey())
		lr.topics = req.GetTopics()
	}

	return lr
}

// log Kafka log records
func (l *kafkaLogRecord) log(verdict accesslog.FlowVerdict, code int, info string) {
	l.ApplyTags(logger.LogTags.Verdict(verdict, info))
	l.Kafka.ErrorCode = code

	// Log multiple entries for multiple Kafka topics in a single request.
	for _, t := range l.topics {
		l.Kafka.Topic.Topic = t
		l.Log()
	}
}

func (k *kafkaRedirect) handleRequest(pair *connectionPair, req *kafka.RequestMessage, correlationCache *kafka.CorrelationCache) {
	scopedLog := log.WithField(fieldID, pair.String())
	flowdebug.Log(scopedLog.WithField(logfields.Request, req.String()), "Handling Kafka request")

	record := k.newLogRecordFromRequest(req)

	addr := pair.Rx.conn.RemoteAddr()
	if addr == nil {
		info := fmt.Sprint("RemoteAddr() is nil")
		scopedLog.Warn(info)
		record.log(accesslog.VerdictError, kafka.ErrInvalidMessage, info)
		return
	}

	// retrieve identity of source together with original destination IP
	// and destination port
	srcIdentity, dstIPPort, err := k.conf.lookupNewDest(addr.String(), k.redirect.ProxyPort)
	if err != nil {
		scopedLog.WithField("source",
			addr.String()).WithError(err).Error("Unable lookup original destination")
		record.log(accesslog.VerdictError, kafka.ErrInvalidMessage,
			fmt.Sprintf("Unable lookup original destination: %s", err))
		return
	}

	record.ApplyTags(logger.LogTags.Addressing(logger.AddressingInfo{
		SrcIPPort:   addr.String(),
		DstIPPort:   dstIPPort,
		SrcIdentity: srcIdentity,
	}))

	if !k.canAccess(req, identity.NumericIdentity(srcIdentity)) {
		flowdebug.Log(scopedLog, "Kafka request is denied by policy")

		record.log(accesslog.VerdictDenied,
			kafka.ErrTopicAuthorizationFailed, fmt.Sprint("Kafka request is denied by policy"))

		resp, err := req.CreateResponse(proto.ErrTopicAuthorizationFailed)
		if err != nil {
			record.log(accesslog.VerdictError,
				kafka.ErrInvalidMessage, fmt.Sprintf("Unable to create response: %s", err))
			scopedLog.WithError(err).Error("Unable to create Kafka response")
			return
		}

		pair.Rx.Enqueue(resp.GetRaw())
		return
	}

	if pair.Tx.Closed() {
		marker := 0
		if !k.conf.noMarker {
			marker = GetMagicMark(k.redirect.ingress, int(srcIdentity))
		}

		flowdebug.Log(scopedLog.WithFields(logrus.Fields{
			"marker":      marker,
			"destination": dstIPPort,
		}), "Dialing original destination")

		txConn, err := ciliumDialer(marker, addr.Network(), dstIPPort)
		if err != nil {
			scopedLog.WithError(err).WithFields(logrus.Fields{
				"origNetwork": addr.Network(),
				"origDest":    dstIPPort,
			}).Error("Unable to dial original destination")

			record.log(accesslog.VerdictError,
				kafka.ErrNetwork, fmt.Sprintf("Unable to dial original destination: %s", err))

			return
		}

		pair.Tx.SetConnection(txConn)

		// Start go routine to handle responses and pass in a copy of
		// the request record as template for all responses
		go k.handleResponseConnection(pair, correlationCache)
	}

	// The request is allowed so we will forward it:
	// 1. Rewrite the correlation ID to a unique ID, it will be restored in
	//    the response direction
	// 2. Store the request in the correlation cache
	correlationCache.HandleRequest(req, nil)

	flowdebug.Log(scopedLog, "Forwarding Kafka request")
	// log valid request
	record.log(accesslog.VerdictForwarded, kafka.ErrNone, "")

	// Write the entire raw request onto the outgoing connection
	pair.Tx.Enqueue(req.GetRaw())
}

type kafkaReqMessageHander func(pair *connectionPair, req *kafka.RequestMessage, correlationCache *kafka.CorrelationCache)
type kafkaRespMessageHander func(pair *connectionPair, req *kafka.ResponseMessage)

func handleRequests(done <-chan struct{}, pair *connectionPair, c *proxyConnection,
	record *kafkaLogRecord, handler kafkaReqMessageHander) {
	defer c.Close()

	// create a correlation cache
	correlationCache := kafka.NewCorrelationCache()
	defer correlationCache.DeleteCache()

	scopedLog := log.WithField(fieldID, pair.String())
	for {
		req, err := kafka.ReadRequest(c.conn)

		// Ignore any error if the listen socket has been closed, i.e. the
		// port redirect has been removed.
		select {
		case <-done:
			scopedLog.Debug("Redirect removed; closing Kafka request connection")
			return
		default:
		}

		if err != nil {
			if record != nil {
				record.log(accesslog.VerdictError,
					kafka.ErrInvalidMessage, fmt.Sprintf("Unable to parse Kafka request: %s", err))
			}
			scopedLog.WithError(err).Error("Unable to parse Kafka request; closing Kafka request connection")
			return
		}

		handler(pair, req, correlationCache)
	}
}

func (k *kafkaRedirect) handleResponses(done <-chan struct{}, pair *connectionPair, c *proxyConnection,
	correlationCache *kafka.CorrelationCache, handler kafkaRespMessageHander) {
	defer c.Close()
	scopedLog := log.WithField(fieldID, pair.String())
	for {
		rsp, err := kafka.ReadResponse(c.conn)

		// Ignore any error if the listen socket has been closed, i.e. the
		// port redirect has been removed.
		select {
		case <-done:
			scopedLog.Debug("Redirect removed; closing Kafka response connection")
			return
		default:
		}

		if err != nil {
			record := k.newLogRecordFromResponse(nil, nil)
			record.log(accesslog.VerdictError,
				kafka.ErrInvalidMessage,
				fmt.Sprintf("Unable to parse Kafka response: %s", err))
			scopedLog.WithError(err).Error("Unable to parse Kafka response; closing Kafka response connection")
			return
		}

		// 1. Find the request that correlates with this response based
		//    on the correlation ID
		// 2. Restore the original correlation id that was overwritten
		//    by the proxy so the client is guaranteed to see the
		//    correlation id as expected
		req := correlationCache.CorrelateResponse(rsp)

		record := k.newLogRecordFromResponse(rsp, req)
		record.log(accesslog.VerdictForwarded, kafka.ErrNone, "")

		handler(pair, rsp)
	}
}

func (k *kafkaRedirect) handleRequestConnection(pair *connectionPair) {
	flowdebug.Log(log.WithFields(logrus.Fields{
		"from": pair.Rx,
		"to":   pair.Tx,
	}), "Proxying request Kafka connection")

	handleRequests(k.socket.closing, pair, pair.Rx, nil, k.handleRequest)

	// The proxymap contains an entry with metadata for the receive side of the
	// connection, remove it after the connection has been closed.
	if pair.Rx != nil {
		// We are running in our own go routine here so we can just
		// block this go routine until after the connection is
		// guaranteed to have been closed
		time.Sleep(proxyConnectionCloseTimeout + time.Second)

		if err := k.redirect.removeProxyMapEntryOnClose(pair.Rx.conn); err != nil {
			log.WithError(err).Warning("Unable to remove proxymap entry after closing connection")
		}
	}
}

func (k *kafkaRedirect) handleResponseConnection(pair *connectionPair, correlationCache *kafka.CorrelationCache) {
	flowdebug.Log(log.WithFields(logrus.Fields{
		"from": pair.Tx,
		"to":   pair.Rx,
	}), "Proxying response Kafka connection")

	k.handleResponses(k.socket.closing, pair, pair.Tx, correlationCache,
		func(pair *connectionPair, rsp *kafka.ResponseMessage) {
			pair.Rx.Enqueue(rsp.GetRaw())
		})
}

// UpdateRules replaces old l7 rules of a redirect with new ones.
func (k *kafkaRedirect) UpdateRules(wg *completion.WaitGroup) error {
	return nil
}

// Close the redirect.
func (k *kafkaRedirect) Close(wg *completion.WaitGroup) {
	k.socket.Close()
}

func init() {
	if err := proto.ConfigureParser(proto.ParserConfig{
		SimplifiedMessageSetParsing: false,
	}); err != nil {
		log.WithError(err).Fatal("Unable to configure kafka parser")
	}
}
