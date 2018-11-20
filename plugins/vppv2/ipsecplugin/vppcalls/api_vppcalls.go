// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vppcalls

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vppv2/model/ipsec"
)

// IPSecVppAPI provides methods for creating and managing of a IPsec configuration
type IPSecVppAPI interface {
	IPSecVppWrite
	IPSecVPPRead
}

// IPSecVppWrite provides write methods for IPsec
type IPSecVppWrite interface {
	// AddTunnelInterface adds tunnel interface
	AddTunnelInterface(tunnel *ipsec.TunnelInterface) (uint32, error)
	// DelTunnelInterface removes tunnel interface
	DelTunnelInterface(ifIdx uint32, tunnel *ipsec.TunnelInterface) error
	// AddSPD adds SPD to VPP via binary API
	AddSPD(spdID uint32) error
	// DelSPD deletes SPD from VPP via binary API
	DeleteSPD(spdID uint32) error
	// InterfaceAddSPD adds SPD interface assignment to VPP via binary API
	InterfaceAddSPD(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// InterfaceDelSPD deletes SPD interface assignment from VPP via binary API
	InterfaceDelSPD(spdID uint32, iface *ipsec.SecurityPolicyDatabase_Interface) error
	// AddSPDEntry adds SPD policy entry to VPP via binary API
	AddSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// DelSPDEntry deletes SPD policy entry from VPP via binary API
	DelSPDEntry(spdID, saID uint32, spd *ipsec.SecurityPolicyDatabase_PolicyEntry) error
	// AddSAEntry adds SA to VPP via binary API
	AddSAEntry(sa *ipsec.SecurityAssociation) error
	// DelSAEntry deletes SA from VPP via binary API
	DeleteSAEntry(sa *ipsec.SecurityAssociation) error
}

// IPSecVPPRead provides read methods for IPSec
type IPSecVPPRead interface {
	// DumpIPSecSPD returns a list of IPSec security policy databases
	DumpIPSecSPD() (spdList []*IPSecSpdDetails, err error)
	// DumpIPSecSA returns a list of configured security associations
	DumpIPSecSA() (saList []*IPSecSaDetails, err error)
	// DumpIPSecSAWithIndex returns a security association with provided index
	DumpIPSecSAWithIndex(saID uint32) (saList []*IPSecSaDetails, err error)
	// DumpIPSecTunnelInterfaces returns a list of configured IPSec tunnel interfaces
	DumpIPSecTunnelInterfaces() (tun []*IPSecTunnelInterfaceDetails, err error)
}

// IPSecVppHandler is accessor for IPsec-related vppcalls methods
type IPSecVppHandler struct {
	callsChannel govppapi.Channel
	ifIndexes    ifaceidx.IfaceMetadataIndex
	log          logging.Logger
}

// NewIPsecVppHandler creates new instance of IPsec vppcalls handler
func NewIPsecVppHandler(callsChan govppapi.Channel, ifIndexes ifaceidx.IfaceMetadataIndex, log logging.Logger) *IPSecVppHandler {
	return &IPSecVppHandler{
		callsChannel: callsChan,
		ifIndexes:    ifIndexes,
		log:          log,
	}
}
