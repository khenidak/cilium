// Copyright 2016-2018 Authors of Cilium
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

package policy

import (
	"strconv"

	"github.com/cilium/cilium/pkg/byteorder"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/u8proto"
)

// SecurityIDContexts maps a security identity to a L4L7Map
type SecurityIDContexts map[identity.NumericIdentity]L4L7Map

// DeepCopy returns a deep copy of SecurityIDContexts
func (sc SecurityIDContexts) DeepCopy() SecurityIDContexts {
	cpy := make(SecurityIDContexts)
	for k, v := range sc {
		cpy[k] = v.DeepCopy()
	}
	return cpy
}

// SecurityIDContexts returns a new L4L7Map created.
func NewSecurityIDContexts() SecurityIDContexts {
	return SecurityIDContexts(make(map[identity.NumericIdentity]L4L7Map))
}

// L4L7Map maps L4 policy-related metadata with L7 policy-related metadata.
type L4L7Map map[L4Rule]L7Rule

// NewL4L7Map returns a new L4L7Map.
func NewL4L7Map() L4L7Map {
	return L4L7Map(make(map[L4Rule]L7Rule))
}

// DeepCopy returns a deep copy of L4L7Map.
func (rc L4L7Map) DeepCopy() L4L7Map {
	cpy := make(L4L7Map)
	for k, v := range rc {
		cpy[k] = v
	}
	return cpy
}

// IsL3Only returns false if the given L4L7Map contains any entry. If it
// does not contain any entry it is considered an L3 only rule.
func (rc L4L7Map) IsL3Only() bool {
	return rc != nil && len(rc) == 0
}

// L4Rule contains the L4-specific parts of a policy rule (port and protocol tuple).
// Do not use pointers for fields in this type since this structure is used as
// a key for maps.
type L4Rule struct {
	// Port is the destination port in the policy in network byte order.
	Port uint16
	// Proto is the protocol ID used.
	Proto uint8
}

// L7Rule contains the L7-specific parts of a policy rule.
type L7Rule struct {
	// RedirectPort is the L7 redirect port in the policy in network byte order.
	RedirectPort uint16
	// L4Installed specifies if the L4 rule is installed in the L4 BPF map.
	L4Installed bool
}

// IsRedirect checks if the L7Rule has a non-zero redirect port. A non-zero
// redirect port means that traffic should be directed to the L7-proxy.
func (rc L7Rule) IsRedirect() bool {
	return rc.RedirectPort != 0
}

// PortProto returns the port proto tuple in a human readable format. i.e.
// with its port in host byte order.
func (rc L4Rule) PortProto() string {
	proto := u8proto.U8proto(rc.Proto).String()
	port := strconv.Itoa(int(byteorder.NetworkToHost(uint16(rc.Port)).(uint16)))
	return port + "/" + proto
}
