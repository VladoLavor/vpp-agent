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

package linuxcalls

import (
	"net"

	"go.ligato.io/vpp-agent/v2/plugins/linux/nsplugin"

	"github.com/ligato/cn-infra/logging"
	"go.ligato.io/vpp-agent/v2/plugins/linux/ifplugin/ifaceidx"
	interfaces "go.ligato.io/vpp-agent/v2/proto/ligato/linux/interfaces"
	namespaces "go.ligato.io/vpp-agent/v2/proto/ligato/linux/namespace"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// InterfaceDetails is an object combining linux interface data based on proto
// model with additional metadata
type InterfaceDetails struct {
	Interface *interfaces.Interface `json:"interface"`
	Meta      *InterfaceMeta        `json:"interface_meta"`
}

// InterfaceMeta represents linux interface metadata
type InterfaceMeta struct {
	LinuxIfIndex int  `json:"linux_if_index"`
	IsExisting   bool `json:"is_existing"`
}

// NetlinkAPI interface covers all methods inside linux calls package
// needed to manage linux interfaces.
type NetlinkAPI interface {
	NetlinkAPIWrite
	NetlinkAPIRead
}

// NetlinkAPIWrite interface covers write methods inside linux calls package
// needed to manage linux interfaces.
type NetlinkAPIWrite interface {
	// AddVethInterfacePair configures two connected VETH interfaces
	AddVethInterfacePair(ifName, peerIfName string) error
	// DeleteInterface removes the given interface.
	DeleteInterface(ifName string) error
	// SetInterfaceUp sets interface state to 'up'
	SetInterfaceUp(ifName string) error
	// SetInterfaceDown sets interface state to 'down'
	SetInterfaceDown(ifName string) error
	// AddInterfaceIP adds new IP address
	AddInterfaceIP(ifName string, addr *net.IPNet) error
	// DelInterfaceIP removes IP address from linux interface
	DelInterfaceIP(ifName string, addr *net.IPNet) error
	// SetInterfaceMac sets MAC address
	SetInterfaceMac(ifName string, macAddress string) error
	// SetInterfaceMTU set maximum transmission unit for interface
	SetInterfaceMTU(ifName string, mtu int) error
	// RenameInterface changes interface host name
	RenameInterface(ifName string, newName string) error
	// SetInterfaceAlias sets the alias of the given interface.
	// Equivalent to: `ip link set dev $ifName alias $alias`
	SetInterfaceAlias(ifName, alias string) error
	// SetLinkNamespace puts link into a network namespace.
	SetLinkNamespace(link netlink.Link, ns netns.NsHandle) error
	// SetChecksumOffloading enables/disables Rx/Tx checksum offloading
	// for the given interface.
	SetChecksumOffloading(ifName string, rxOn, txOn bool) error
}

// NetlinkAPIRead interface covers read methods inside linux calls package
// needed to manage linux interfaces.
type NetlinkAPIRead interface {
	// GetLinkByName returns netlink interface type
	GetLinkByName(ifName string) (netlink.Link, error)
	// GetLinkList return all links from namespace
	GetLinkList() ([]netlink.Link, error)
	// LinkSubscribe takes a channel to which notifications will be sent
	// when links change. Close the 'done' chan to stop subscription.
	LinkSubscribe(ch chan<- netlink.LinkUpdate, done <-chan struct{}) error
	// GetAddressList reads all IP addresses
	GetAddressList(ifName string) ([]netlink.Addr, error)
	// InterfaceExists verifies interface existence
	InterfaceExists(ifName string) (bool, error)
	// IsInterfaceUp checks if the interface is UP.
	IsInterfaceUp(ifName string) (bool, error)
	// GetInterfaceType returns linux interface type
	GetInterfaceType(ifName string) (string, error)
	// GetChecksumOffloading returns the state of Rx/Tx checksum offloading
	// for the given interface.
	GetChecksumOffloading(ifName string) (rxOn, txOn bool, err error)
	// DumpInterfaces uses local cache to gather information about linux
	// namespaces and retrieves them.
	DumpInterfaces() ([]*InterfaceDetails, error)
	// DumpInterfacesWithContext retrieves all linux interfaces based
	// on provided namespace context.
	DumpInterfacesWithContext(nsList []*namespaces.NetNamespace) ([]*InterfaceDetails, error)
}

// NetLinkHandler is accessor for Netlink methods.
type NetLinkHandler struct {
	nsPlugin    nsplugin.API
	ifIndexes   ifaceidx.LinuxIfMetadataIndex
	agentPrefix string

	// parallelization of the Retrieve operation
	goRoutineCount int

	log logging.Logger
}

// NewNetLinkHandler creates new instance of Netlink handler.
func NewNetLinkHandler(nsPlugin nsplugin.API, ifIndexes ifaceidx.LinuxIfMetadataIndex, agentPrefix string,
	goRoutineCount int, log logging.Logger) *NetLinkHandler {
	return &NetLinkHandler{
		nsPlugin:       nsPlugin,
		ifIndexes:      ifIndexes,
		agentPrefix:    agentPrefix,
		goRoutineCount: goRoutineCount,
		log:            log,
	}
}
