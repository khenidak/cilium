// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/api/v2/lds.proto

package v2

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import envoy_api_v2_core "github.com/cilium/cilium/pkg/envoy/envoy/api/v2/core"
import envoy_api_v2_core1 "github.com/cilium/cilium/pkg/envoy/envoy/api/v2/core"
import envoy_api_v2_listener "github.com/cilium/cilium/pkg/envoy/envoy/api/v2/listener"
import _ "google.golang.org/genproto/googleapis/api/annotations"
import google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
import _ "github.com/lyft/protoc-gen-validate/validate"
import _ "github.com/gogo/protobuf/gogoproto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Listener_DrainType int32

const (
	// Drain in response to calling /healthcheck/fail admin endpoint (along with the health check
	// filter), listener removal/modification, and hot restart.
	Listener_DEFAULT Listener_DrainType = 0
	// Drain in response to listener removal/modification and hot restart. This setting does not
	// include /healthcheck/fail. This setting may be desirable if Envoy is hosting both ingress
	// and egress listeners.
	Listener_MODIFY_ONLY Listener_DrainType = 1
)

var Listener_DrainType_name = map[int32]string{
	0: "DEFAULT",
	1: "MODIFY_ONLY",
}
var Listener_DrainType_value = map[string]int32{
	"DEFAULT":     0,
	"MODIFY_ONLY": 1,
}

func (x Listener_DrainType) String() string {
	return proto.EnumName(Listener_DrainType_name, int32(x))
}
func (Listener_DrainType) EnumDescriptor() ([]byte, []int) { return fileDescriptor3, []int{0, 0} }

type Listener struct {
	// The unique name by which this listener is known. If no name is provided,
	// Envoy will allocate an internal UUID for the listener. If the listener is to be dynamically
	// updated or removed via :ref:`LDS <config_listeners_lds>` a unique name must be provided.
	// By default, the maximum length of a listener's name is limited to 60 characters. This limit can
	// be increased by setting the :option:`--max-obj-name-len` command line argument to the desired
	// value.
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	// The address that the listener should listen on. In general, the address must be unique, though
	// that is governed by the bind rules of the OS. E.g., multiple listeners can listen on port 0 on
	// Linux as the actual port will be allocated by the OS.
	Address *envoy_api_v2_core.Address `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
	// A list of filter chains to consider for this listener. The
	// :ref:`FilterChain <envoy_api_msg_listener.FilterChain>` with the most specific
	// :ref:`FilterChainMatch <envoy_api_msg_listener.FilterChainMatch>` criteria is used on a
	// connection.
	//
	// .. attention::
	//
	//   In the current version, multiple filter chains are supported **only** so that SNI can be
	//   configured. See the :ref:`FAQ entry <faq_how_to_setup_sni>` on how to configure SNI for more
	//   information. When multiple filter chains are configured, each filter chain must have an
	//   **identical** set of :ref:`filters <envoy_api_field_listener.FilterChain.filters>`. If the
	//   filters differ, the configuration will fail to load. In the future, this limitation will be
	//   relaxed such that different filters can be used depending on which filter chain matches
	//   (based on SNI or some other parameter).
	FilterChains []*envoy_api_v2_listener.FilterChain `protobuf:"bytes,3,rep,name=filter_chains,json=filterChains" json:"filter_chains,omitempty"`
	// If a connection is redirected using *iptables*, the port on which the proxy
	// receives it might be different from the original destination address. When this flag is set to
	// true, the listener hands off redirected connections to the listener associated with the
	// original destination address. If there is no listener associated with the original destination
	// address, the connection is handled by the listener that receives it. Defaults to false.
	//
	// .. attention::
	//
	//   This field is deprecated. Use :ref:`an original_dst <config_listener_filters_original_dst>`
	//   :ref:`listener filter <envoy_api_field_Listener.listener_filters>` instead.
	//
	//   Note that hand off to another listener is *NOT* performed without this flag. Once
	//   :ref:`FilterChainMatch <envoy_api_msg_listener.FilterChainMatch>` is implemented this flag
	//   will be removed, as filter chain matching can be used to select a filter chain based on the
	//   restored destination address.
	UseOriginalDst *google_protobuf.BoolValue `protobuf:"bytes,4,opt,name=use_original_dst,json=useOriginalDst" json:"use_original_dst,omitempty"`
	// Soft limit on size of the listener’s new connection read and write buffers.
	// If unspecified, an implementation defined default is applied (1MiB).
	PerConnectionBufferLimitBytes *google_protobuf.UInt32Value `protobuf:"bytes,5,opt,name=per_connection_buffer_limit_bytes,json=perConnectionBufferLimitBytes" json:"per_connection_buffer_limit_bytes,omitempty"`
	// Listener metadata.
	Metadata *envoy_api_v2_core1.Metadata `protobuf:"bytes,6,opt,name=metadata" json:"metadata,omitempty"`
	// [#not-implemented-hide:]
	DeprecatedV1 *Listener_DeprecatedV1 `protobuf:"bytes,7,opt,name=deprecated_v1,json=deprecatedV1" json:"deprecated_v1,omitempty"`
	// The type of draining to perform at a listener-wide level.
	DrainType Listener_DrainType `protobuf:"varint,8,opt,name=drain_type,json=drainType,enum=envoy.api.v2.Listener_DrainType" json:"drain_type,omitempty"`
	// Listener filters have the opportunity to manipulate and augment the connection metadata that
	// is used in connection filter chain matching, for example. These filters are run before any in
	// :ref:`filter_chains <envoy_api_field_Listener.filter_chains>`. Order matters as the
	// filters are processed sequentially right after a socket has been accepted by the listener, and
	// before a connection is created.
	ListenerFilters []*envoy_api_v2_listener.ListenerFilter `protobuf:"bytes,9,rep,name=listener_filters,json=listenerFilters" json:"listener_filters,omitempty"`
	// [#not-implemented-hide:]
	// Whether the listener should be set as a transparent socket.
	// When this flag is set to true, connections can be redirected to the listener using an
	// *iptables* *TPROXY* target, in which case the original source and destination addresses and
	// ports are preserved on accepted connections. This flag should be used in combination with
	// :ref:`an original_dst <config_listener_filters_original_dst>` :ref:`listener filter
	// <envoy_api_field_Listener.listener_filters>` to mark the connections' local addresses as
	// "restored." This can be used to hand off each redirected connection to another listener
	// associated with the connection's destination address. Direct connections to the socket without
	// using *TPROXY* cannot be distinguished from connections redirected using *TPROXY* and are
	// therefore treated as if they were redirected.
	// When this flag is set to false, the listener's socket is explicitly reset as non-transparent.
	// Setting this flag requires Envoy to run with the *CAP_NET_ADMIN* capability.
	// When this flag is not set (default), the socket is not modified, i.e. the transparent option
	// is neither set nor reset.
	Transparent *google_protobuf.BoolValue `protobuf:"bytes,10,opt,name=transparent" json:"transparent,omitempty"`
	// [#not-implemented-hide:] Whether the listener should set the IP_FREEBIND socket option. When
	// this flag is set to true, listeners can be bound to an IP address that is not configured on
	// the system running Envoy.
	// When this flag is set to false, the option IP_FREEBIND is disabled on the socket.
	// When this flag is not set (default), the socket is not modified, i.e. the option is neither
	// enabled nor disabled.
	Freebind *google_protobuf.BoolValue `protobuf:"bytes,11,opt,name=freebind" json:"freebind,omitempty"`
}

func (m *Listener) Reset()                    { *m = Listener{} }
func (m *Listener) String() string            { return proto.CompactTextString(m) }
func (*Listener) ProtoMessage()               {}
func (*Listener) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *Listener) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Listener) GetAddress() *envoy_api_v2_core.Address {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Listener) GetFilterChains() []*envoy_api_v2_listener.FilterChain {
	if m != nil {
		return m.FilterChains
	}
	return nil
}

func (m *Listener) GetUseOriginalDst() *google_protobuf.BoolValue {
	if m != nil {
		return m.UseOriginalDst
	}
	return nil
}

func (m *Listener) GetPerConnectionBufferLimitBytes() *google_protobuf.UInt32Value {
	if m != nil {
		return m.PerConnectionBufferLimitBytes
	}
	return nil
}

func (m *Listener) GetMetadata() *envoy_api_v2_core1.Metadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *Listener) GetDeprecatedV1() *Listener_DeprecatedV1 {
	if m != nil {
		return m.DeprecatedV1
	}
	return nil
}

func (m *Listener) GetDrainType() Listener_DrainType {
	if m != nil {
		return m.DrainType
	}
	return Listener_DEFAULT
}

func (m *Listener) GetListenerFilters() []*envoy_api_v2_listener.ListenerFilter {
	if m != nil {
		return m.ListenerFilters
	}
	return nil
}

func (m *Listener) GetTransparent() *google_protobuf.BoolValue {
	if m != nil {
		return m.Transparent
	}
	return nil
}

func (m *Listener) GetFreebind() *google_protobuf.BoolValue {
	if m != nil {
		return m.Freebind
	}
	return nil
}

// [#not-implemented-hide:]
type Listener_DeprecatedV1 struct {
	// Whether the listener should bind to the port. A listener that doesn’t
	// bind can only receive connections redirected from other listeners that
	// set use_original_dst parameter to true. Default is true.
	//
	// [V2-API-DIFF] This is deprecated in v2, all Listeners will bind to their
	// port. An additional filter chain must be created for every original
	// destination port this listener may redirect to in v2, with the original
	// port specified in the FilterChainMatch destination_port field.
	BindToPort *google_protobuf.BoolValue `protobuf:"bytes,1,opt,name=bind_to_port,json=bindToPort" json:"bind_to_port,omitempty"`
}

func (m *Listener_DeprecatedV1) Reset()                    { *m = Listener_DeprecatedV1{} }
func (m *Listener_DeprecatedV1) String() string            { return proto.CompactTextString(m) }
func (*Listener_DeprecatedV1) ProtoMessage()               {}
func (*Listener_DeprecatedV1) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0, 0} }

func (m *Listener_DeprecatedV1) GetBindToPort() *google_protobuf.BoolValue {
	if m != nil {
		return m.BindToPort
	}
	return nil
}

func init() {
	proto.RegisterType((*Listener)(nil), "envoy.api.v2.Listener")
	proto.RegisterType((*Listener_DeprecatedV1)(nil), "envoy.api.v2.Listener.DeprecatedV1")
	proto.RegisterEnum("envoy.api.v2.Listener_DrainType", Listener_DrainType_name, Listener_DrainType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ListenerDiscoveryService service

type ListenerDiscoveryServiceClient interface {
	StreamListeners(ctx context.Context, opts ...grpc.CallOption) (ListenerDiscoveryService_StreamListenersClient, error)
	FetchListeners(ctx context.Context, in *DiscoveryRequest, opts ...grpc.CallOption) (*DiscoveryResponse, error)
}

type listenerDiscoveryServiceClient struct {
	cc *grpc.ClientConn
}

func NewListenerDiscoveryServiceClient(cc *grpc.ClientConn) ListenerDiscoveryServiceClient {
	return &listenerDiscoveryServiceClient{cc}
}

func (c *listenerDiscoveryServiceClient) StreamListeners(ctx context.Context, opts ...grpc.CallOption) (ListenerDiscoveryService_StreamListenersClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ListenerDiscoveryService_serviceDesc.Streams[0], c.cc, "/envoy.api.v2.ListenerDiscoveryService/StreamListeners", opts...)
	if err != nil {
		return nil, err
	}
	x := &listenerDiscoveryServiceStreamListenersClient{stream}
	return x, nil
}

type ListenerDiscoveryService_StreamListenersClient interface {
	Send(*DiscoveryRequest) error
	Recv() (*DiscoveryResponse, error)
	grpc.ClientStream
}

type listenerDiscoveryServiceStreamListenersClient struct {
	grpc.ClientStream
}

func (x *listenerDiscoveryServiceStreamListenersClient) Send(m *DiscoveryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *listenerDiscoveryServiceStreamListenersClient) Recv() (*DiscoveryResponse, error) {
	m := new(DiscoveryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *listenerDiscoveryServiceClient) FetchListeners(ctx context.Context, in *DiscoveryRequest, opts ...grpc.CallOption) (*DiscoveryResponse, error) {
	out := new(DiscoveryResponse)
	err := grpc.Invoke(ctx, "/envoy.api.v2.ListenerDiscoveryService/FetchListeners", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ListenerDiscoveryService service

type ListenerDiscoveryServiceServer interface {
	StreamListeners(ListenerDiscoveryService_StreamListenersServer) error
	FetchListeners(context.Context, *DiscoveryRequest) (*DiscoveryResponse, error)
}

func RegisterListenerDiscoveryServiceServer(s *grpc.Server, srv ListenerDiscoveryServiceServer) {
	s.RegisterService(&_ListenerDiscoveryService_serviceDesc, srv)
}

func _ListenerDiscoveryService_StreamListeners_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ListenerDiscoveryServiceServer).StreamListeners(&listenerDiscoveryServiceStreamListenersServer{stream})
}

type ListenerDiscoveryService_StreamListenersServer interface {
	Send(*DiscoveryResponse) error
	Recv() (*DiscoveryRequest, error)
	grpc.ServerStream
}

type listenerDiscoveryServiceStreamListenersServer struct {
	grpc.ServerStream
}

func (x *listenerDiscoveryServiceStreamListenersServer) Send(m *DiscoveryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *listenerDiscoveryServiceStreamListenersServer) Recv() (*DiscoveryRequest, error) {
	m := new(DiscoveryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ListenerDiscoveryService_FetchListeners_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiscoveryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ListenerDiscoveryServiceServer).FetchListeners(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/envoy.api.v2.ListenerDiscoveryService/FetchListeners",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ListenerDiscoveryServiceServer).FetchListeners(ctx, req.(*DiscoveryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ListenerDiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "envoy.api.v2.ListenerDiscoveryService",
	HandlerType: (*ListenerDiscoveryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchListeners",
			Handler:    _ListenerDiscoveryService_FetchListeners_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamListeners",
			Handler:       _ListenerDiscoveryService_StreamListeners_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "envoy/api/v2/lds.proto",
}

func init() { proto.RegisterFile("envoy/api/v2/lds.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 704 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0xbf, 0x6e, 0x13, 0x4b,
	0x14, 0xc6, 0x33, 0xce, 0x1f, 0xdb, 0x63, 0xc7, 0xb1, 0x46, 0x57, 0x37, 0x2b, 0xdf, 0xdc, 0xc4,
	0xd7, 0x17, 0x24, 0x43, 0xb1, 0x26, 0x8e, 0x04, 0x52, 0x14, 0x09, 0xc5, 0x31, 0x56, 0x22, 0x39,
	0x04, 0x6d, 0xfe, 0x40, 0xaa, 0xd5, 0xd8, 0x7b, 0xec, 0x8c, 0xb4, 0x9e, 0x59, 0x66, 0xc6, 0x46,
	0x6e, 0xa9, 0x10, 0x25, 0xbc, 0x04, 0xcf, 0x40, 0x45, 0x49, 0x4f, 0x4f, 0x81, 0x68, 0x10, 0x6f,
	0x40, 0x85, 0x76, 0xbd, 0xb3, 0xd8, 0x4a, 0x42, 0x1a, 0xba, 0xb3, 0xe7, 0x7c, 0xe7, 0x37, 0xa3,
	0xef, 0xd3, 0x0e, 0xfe, 0x1b, 0xf8, 0x48, 0x8c, 0x6b, 0x34, 0x60, 0xb5, 0x51, 0xbd, 0xe6, 0x7b,
	0xca, 0x0e, 0xa4, 0xd0, 0x82, 0xe4, 0xa3, 0xbe, 0x4d, 0x03, 0x66, 0x8f, 0xea, 0xa5, 0x8d, 0x19,
	0x55, 0x57, 0x48, 0xa8, 0x51, 0xcf, 0x93, 0xa0, 0x62, 0x79, 0x69, 0xed, 0xb2, 0xa0, 0x43, 0x15,
	0x5c, 0x39, 0xf5, 0x98, 0xea, 0x8a, 0x11, 0xc8, 0x71, 0x3c, 0xbd, 0x35, 0x7b, 0x05, 0xa6, 0x34,
	0x70, 0x90, 0x49, 0x61, 0x18, 0x7d, 0x21, 0xfa, 0x3e, 0x44, 0x32, 0xca, 0xb9, 0xd0, 0x54, 0x33,
	0xc1, 0xcd, 0xf9, 0xeb, 0xf1, 0x34, 0xfa, 0xea, 0x0c, 0x7b, 0xb5, 0x17, 0x92, 0x06, 0x01, 0x48,
	0x33, 0x5f, 0x1d, 0x51, 0x9f, 0x79, 0x54, 0x43, 0xcd, 0x14, 0xf1, 0xe0, 0xaf, 0xbe, 0xe8, 0x8b,
	0xa8, 0xac, 0x85, 0xd5, 0xa4, 0x5b, 0xf9, 0xb1, 0x84, 0x33, 0xed, 0xf8, 0x7c, 0x42, 0xf0, 0x02,
	0xa7, 0x03, 0xb0, 0x50, 0x19, 0x55, 0xb3, 0x4e, 0x54, 0x93, 0x26, 0x4e, 0xc7, 0x06, 0x58, 0xa9,
	0x32, 0xaa, 0xe6, 0xea, 0x25, 0x7b, 0xda, 0x30, 0x3b, 0x74, 0xc0, 0xde, 0x9d, 0x28, 0x1a, 0x85,
	0x8f, 0x9f, 0x37, 0xe6, 0xde, 0x7f, 0xfb, 0x30, 0xbf, 0xf8, 0x1a, 0xa5, 0x8a, 0xc8, 0x31, 0xab,
	0xe4, 0x29, 0x5e, 0xee, 0x31, 0x5f, 0x83, 0x74, 0xbb, 0x17, 0x94, 0x71, 0x65, 0xcd, 0x97, 0xe7,
	0xab, 0xb9, 0x7a, 0x65, 0x96, 0x95, 0x18, 0xd1, 0x8a, 0xb4, 0x7b, 0xa1, 0x74, 0x8a, 0xf9, 0x06,
	0xa5, 0x32, 0xc8, 0xc9, 0xf7, 0x7e, 0x0d, 0x15, 0xd9, 0xc7, 0xc5, 0xa1, 0x02, 0x57, 0x48, 0xd6,
	0x67, 0x9c, 0xfa, 0xae, 0xa7, 0xb4, 0xb5, 0x10, 0xdf, 0x73, 0xe2, 0x94, 0x6d, 0x9c, 0xb2, 0x1b,
	0x42, 0xf8, 0x67, 0xd4, 0x1f, 0x42, 0x23, 0x65, 0x21, 0xa7, 0x30, 0x54, 0x70, 0x14, 0xaf, 0x35,
	0x95, 0x26, 0x3d, 0xfc, 0x5f, 0x10, 0xde, 0x4f, 0x70, 0x0e, 0xdd, 0xd0, 0x71, 0xb7, 0x33, 0xec,
	0xf5, 0x40, 0xba, 0x3e, 0x1b, 0x30, 0xed, 0x76, 0xc6, 0x1a, 0x94, 0xb5, 0x18, 0xa1, 0xd7, 0x2e,
	0xa1, 0x4f, 0x0f, 0xb8, 0xde, 0xaa, 0x47, 0x70, 0xe7, 0xdf, 0x00, 0xe4, 0x5e, 0x42, 0x69, 0x44,
	0x90, 0x76, 0xc8, 0x68, 0x84, 0x08, 0xf2, 0x00, 0x67, 0x06, 0xa0, 0xa9, 0x47, 0x35, 0xb5, 0x96,
	0x22, 0xdc, 0x3f, 0x57, 0x38, 0x7a, 0x18, 0x4b, 0x9c, 0x44, 0x4c, 0xf6, 0xf1, 0xb2, 0x07, 0x81,
	0x84, 0x2e, 0xd5, 0xe0, 0xb9, 0xa3, 0x4d, 0x2b, 0x1d, 0x6d, 0xff, 0x3f, 0xbb, 0x6d, 0xc2, 0xb4,
	0x9b, 0x89, 0xf6, 0x6c, 0xd3, 0xc9, 0x7b, 0x53, 0x5f, 0xe4, 0x21, 0xc6, 0x9e, 0xa4, 0x8c, 0xbb,
	0x7a, 0x1c, 0x80, 0x95, 0x29, 0xa3, 0x6a, 0xa1, 0x5e, 0xbe, 0x0e, 0x13, 0x0a, 0x4f, 0xc6, 0x01,
	0x38, 0x59, 0xcf, 0x94, 0xe4, 0x0c, 0x17, 0x4d, 0x56, 0xee, 0x24, 0x0e, 0x65, 0x65, 0xa3, 0x44,
	0x6f, 0x5f, 0x93, 0xa8, 0xe1, 0x4d, 0x92, 0x6d, 0x2c, 0x84, 0xa1, 0x3a, 0x2b, 0xfe, 0x4c, 0x57,
	0x91, 0x1d, 0x9c, 0xd3, 0x92, 0x72, 0x15, 0x50, 0x09, 0x5c, 0x5b, 0xf8, 0xa6, 0x20, 0x9d, 0x69,
	0x39, 0xb9, 0x8f, 0x33, 0x3d, 0x09, 0xd0, 0x61, 0xdc, 0xb3, 0x72, 0x37, 0xae, 0x26, 0xda, 0x52,
	0x1b, 0xe7, 0xa7, 0xcd, 0x22, 0x3b, 0x38, 0x1f, 0xf6, 0x5d, 0x2d, 0xdc, 0x40, 0x48, 0x1d, 0xfd,
	0x0e, 0xbf, 0x67, 0xe1, 0x50, 0x7f, 0x22, 0x9e, 0x08, 0xa9, 0x2b, 0x77, 0x70, 0x36, 0xf1, 0x8c,
	0xe4, 0x70, 0xba, 0xf9, 0xa8, 0xb5, 0x7b, 0xda, 0x3e, 0x29, 0xce, 0x91, 0x15, 0x9c, 0x3b, 0x3c,
	0x6a, 0x1e, 0xb4, 0xce, 0xdd, 0xa3, 0xc7, 0xed, 0xf3, 0x22, 0xaa, 0x7f, 0x47, 0xd8, 0x32, 0xc6,
	0x34, 0xcd, 0x5b, 0x71, 0x0c, 0x72, 0xc4, 0xba, 0x40, 0x9e, 0xe1, 0x95, 0x63, 0x2d, 0x81, 0x0e,
	0x8c, 0x42, 0x91, 0xf5, 0x59, 0x73, 0x93, 0x15, 0x07, 0x9e, 0x0f, 0x41, 0xe9, 0xd2, 0xc6, 0xb5,
	0x73, 0x15, 0x08, 0xae, 0xa0, 0x32, 0x57, 0x45, 0xf7, 0x10, 0x19, 0xe2, 0x42, 0x0b, 0x74, 0xf7,
	0xe2, 0x0f, 0x82, 0x2b, 0x2f, 0x3f, 0x7d, 0x7d, 0x9b, 0x5a, 0xab, 0xac, 0xce, 0x3c, 0x7b, 0xdb,
	0x26, 0x5f, 0xb5, 0x8d, 0xee, 0x36, 0xd2, 0xef, 0xbe, 0xac, 0xa3, 0x57, 0x08, 0x75, 0x96, 0x22,
	0x07, 0xb7, 0x7e, 0x06, 0x00, 0x00, 0xff, 0xff, 0x97, 0x6c, 0xa0, 0x6d, 0x92, 0x05, 0x00, 0x00,
}
