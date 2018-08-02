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
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	bin_api "github.com/ligato/vpp-agent/plugins/vpp/binapi/nat"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/vpp/model/nat"
)

func (handler *natVppHandler) Nat44GlobalConfigDump(swIfIndices ifaceidx.SwIfIndex) (*nat.Nat44Global, error) {
	handler.log.Debug("dumping Nat44Global")
	// Dump all necessary data to reconstruct global NAT configuration
	isEnabled, err := handler.nat44IsForwardingEnabled()
	if err != nil {
		return nil, err
	}
	natInterfaces, err := handler.Nat44InterfaceDump(swIfIndices)
	if err != nil {
		return nil, err
	}
	natOutputFeature, err := handler.nat44InterfaceOutputFeatureDump(swIfIndices)
	if err != nil {
		return nil, err
	}
	natAddressPools, err := handler.nat44AddressDump()
	if err != nil {
		return nil, err
	}

	// Combine interfaces with output feature with the rest of them
	var nat44GlobalInterfaces []*nat.Nat44Global_NatInterface
	for _, natInterface := range natInterfaces {
		nat44GlobalInterfaces = append(nat44GlobalInterfaces, &nat.Nat44Global_NatInterface{
			Name:          natInterface.Name,
			IsInside:      natInterface.IsInside,
			OutputFeature: false,
		})
	}
	for _, natInterface := range natOutputFeature {
		nat44GlobalInterfaces = append(nat44GlobalInterfaces, &nat.Nat44Global_NatInterface{
			Name:          natInterface.Name,
			IsInside:      natInterface.IsInside,
			OutputFeature: true,
		})
	}

	handler.log.Debug("dumped Nat44Global")

	// Set fields
	return &nat.Nat44Global{
		Forwarding:    isEnabled,
		NatInterfaces: nat44GlobalInterfaces,
		AddressPools:  natAddressPools,
	}, nil
}

func (handler *natVppHandler) NAT44DNatDump(swIfIndices ifaceidx.SwIfIndex) (*nat.Nat44DNat, error) {
	// List od DNAT configs
	var dNatCfgs []*nat.Nat44DNat_DNatConfig

	handler.log.Debug("dumping DNat")

	// Static mappings
	natStMappings, err := handler.nat44StaticMappingDump(swIfIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings: %v", err)
	}
	for tag, data := range natStMappings {
		handler.processDNatData(tag, data, &dNatCfgs)
	}
	// Static mappings with load balancer
	natStLbMappings, err := handler.nat44StaticMappingLbDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 static mappings with load balancer: %v", err)
	}
	for tag, data := range natStLbMappings {
		handler.processDNatData(tag, data, &dNatCfgs)
	}
	// Identity mappings
	natIdMappings, err := handler.nat44IdentityMappingDump(swIfIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to dump NAT44 identity mappings: %v", err)
	}
	for tag, data := range natIdMappings {
		handler.processDNatData(tag, data, &dNatCfgs)
	}

	handler.log.Debugf("dumped %d NAT44DNat configs", len(dNatCfgs))

	return &nat.Nat44DNat{
		DnatConfigs: dNatCfgs,
	}, nil
}

// nat44AddressDump returns a list of NAT44 address pools configured in the VPP
func (handler *natVppHandler) nat44AddressDump() (addresses []*nat.Nat44Global_AddressPool, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44AddressDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44AddressDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44AddressDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 Address pool: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

		addresses = append(addresses, &nat.Nat44Global_AddressPool{
			FirstSrcAddress: ipAddress.To4().String(),
			VrfId:           msg.VrfID,
			TwiceNat:        uintToBool(msg.TwiceNat),
		})
	}

	handler.log.Debugf("NAT44 address pool dump complete, found %d entries", len(addresses))

	return
}

// nat44StaticMappingDump returns a map of static mapping tag/data pairs
func (handler *natVppHandler) nat44StaticMappingDump(swIfIndices ifaceidx.SwIfIndex) (entries map[string]*nat.Nat44DNat_DNatConfig_StaticMapping, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44StaticMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	entries = make(map[string]*nat.Nat44DNat_DNatConfig_StaticMapping)
	req := &bin_api.Nat44StaticMappingDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44StaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 static mapping: %v", err)
		}
		if stop {
			break
		}
		var locals []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP
		lcIPAddress := net.IP(msg.LocalIPAddress)
		exIPAddress := net.IP(msg.ExternalIPAddress)

		// Parse tag (key)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Fill data (value)
		entries[tag] = &nat.Nat44DNat_DNatConfig_StaticMapping{
			VrfId: msg.VrfID,
			ExternalInterface: func(ifIdx uint32) string {
				ifName, _, found := swIfIndices.LookupName(ifIdx)
				if !found && ifIdx != 0xffffffff {
					handler.log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.ExternalSwIfIndex),
			ExternalIp:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps: append(locals, &nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{ // single-value
				LocalIp:   lcIPAddress.To4().String(),
				LocalPort: uint32(msg.LocalPort),
			}),
			Protocol: handler.getNatProtocol(msg.Protocol),
			TwiceNat: handler.getTwiceNatMode(msg.TwiceNat, msg.SelfTwiceNat),
		}
	}

	handler.log.Debugf("NAT44 static mapping dump complete, found %d entries", len(entries))

	return entries, nil
}

// nat44StaticMappingLbDump returns a map of static mapping tag/data pairs with load balancer
func (handler *natVppHandler) nat44StaticMappingLbDump() (entries map[string]*nat.Nat44DNat_DNatConfig_StaticMapping, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44LbStaticMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	entries = make(map[string]*nat.Nat44DNat_DNatConfig_StaticMapping)
	req := &bin_api.Nat44LbStaticMappingDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44LbStaticMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 lb-static mapping: %v", err)
		}
		if stop {
			break
		}

		// Parse tag (key)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Prepare localIPs
		var locals []*nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP
		for _, localIPVal := range msg.Locals {
			localIP := net.IP(localIPVal.Addr)
			locals = append(locals, &nat.Nat44DNat_DNatConfig_StaticMapping_LocalIP{
				LocalIp:     localIP.To4().String(),
				LocalPort:   uint32(localIPVal.Port),
				Probability: uint32(localIPVal.Probability),
			})
		}
		exIPAddress := net.IP(msg.ExternalAddr)

		entries[tag] = &nat.Nat44DNat_DNatConfig_StaticMapping{
			VrfId:        msg.VrfID,
			ExternalIp:   exIPAddress.To4().String(),
			ExternalPort: uint32(msg.ExternalPort),
			LocalIps:     locals,
			Protocol:     handler.getNatProtocol(msg.Protocol),
			TwiceNat:     handler.getTwiceNatMode(msg.TwiceNat, msg.SelfTwiceNat),
		}
	}

	handler.log.Debugf("NAT44 lb-static mapping dump complete, found %d entries", len(entries))

	return entries, nil
}

// nat44IdentityMappingDump returns a map of identity mapping tag/data pairs
func (handler *natVppHandler) nat44IdentityMappingDump(swIfIndices ifaceidx.SwIfIndex) (entries map[string]*nat.Nat44DNat_DNatConfig_IdentityMapping, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44IdentityMappingDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	entries = make(map[string]*nat.Nat44DNat_DNatConfig_IdentityMapping)
	req := &bin_api.Nat44IdentityMappingDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44IdentityMappingDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 identity mapping: %v", err)
		}
		if stop {
			break
		}

		ipAddress := net.IP(msg.IPAddress)

		// Parse tag (key)
		tag := string(bytes.SplitN(msg.Tag, []byte{0x00}, 2)[0])

		// Fill data (value)
		entries[tag] = &nat.Nat44DNat_DNatConfig_IdentityMapping{
			VrfId: msg.VrfID,
			AddressedInterface: func(ifIdx uint32) string {
				ifName, _, found := swIfIndices.LookupName(ifIdx)
				if !found && ifIdx != 0xffffffff {
					handler.log.Warnf("Interface with index %v not found in the mapping", ifIdx)
				}
				return ifName
			}(msg.SwIfIndex),
			IpAddress: ipAddress.To4().String(),
			Port:      uint32(msg.Port),
			Protocol:  handler.getNatProtocol(msg.Protocol),
		}
	}

	handler.log.Debugf("NAT44 identity mapping dump complete, found %d entries", len(entries))

	return entries, nil
}

func (handler *natVppHandler) Nat44InterfaceDump(swIfIndices ifaceidx.SwIfIndex) (interfaces []*nat.Nat44Global_NatInterface, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44InterfaceDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44InterfaceDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := swIfIndices.LookupName(msg.SwIfIndex)
		if !found {
			handler.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		if msg.IsInside == 0 || msg.IsInside == 2 {
			interfaces = append(interfaces, &nat.Nat44Global_NatInterface{
				Name:     ifName,
				IsInside: false,
			})
		}
		if msg.IsInside == 1 || msg.IsInside == 2 {
			interfaces = append(interfaces, &nat.Nat44Global_NatInterface{
				Name:     ifName,
				IsInside: true,
			})
		}
	}

	handler.log.Debugf("NAT44 interface dump complete, found %d entries", len(interfaces))

	return
}

// nat44InterfaceOutputFeatureDump returns a list of interfaces with output feature set
func (handler *natVppHandler) nat44InterfaceOutputFeatureDump(swIfIndices ifaceidx.SwIfIndex) (ifaces []*nat.Nat44Global_NatInterface, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44InterfaceOutputFeatureDump{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44InterfaceOutputFeatureDump{}
	reqContext := handler.dumpChannel.SendMultiRequest(req)

	for {
		msg := &bin_api.Nat44InterfaceOutputFeatureDetails{}
		stop, err := reqContext.ReceiveReply(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to dump NAT44 interface output feature: %v", err)
		}
		if stop {
			break
		}

		// Find interface name
		ifName, _, found := swIfIndices.LookupName(msg.SwIfIndex)
		if !found {
			handler.log.Warnf("Interface with index %d not found in the mapping", msg.SwIfIndex)
			continue
		}

		ifaces = append(ifaces, &nat.Nat44Global_NatInterface{
			Name:          ifName,
			IsInside:      uintToBool(msg.IsInside),
			OutputFeature: true,
		})
	}

	handler.log.Debugf("NAT44 interface with output feature dump complete, found %d entries", len(ifaces))

	return ifaces, nil
}

// Nat44IsForwardingEnabled returns a list of interfaces enabled for NAT44
func (handler *natVppHandler) nat44IsForwardingEnabled() (isEnabled bool, err error) {
	defer func(t time.Time) {
		handler.stopwatch.TimeLog(bin_api.Nat44ForwardingIsEnabled{}).LogTimeEntry(time.Since(t))
	}(time.Now())

	req := &bin_api.Nat44ForwardingIsEnabled{}

	reply := &bin_api.Nat44ForwardingIsEnabledReply{}
	if err := handler.dumpChannel.SendRequest(req).ReceiveReply(reply); err != nil {
		return false, fmt.Errorf("failed to dump forwarding: %v", err)
	}

	isEnabled = uintToBool(reply.Enabled)
	handler.log.Debugf("NAT44 forwarding dump complete, is enabled: %v", isEnabled)

	return isEnabled, nil
}

// Common function can process all static and identity mappings
func (handler *natVppHandler) processDNatData(tag string, data interface{}, dNatCfgs *[]*nat.Nat44DNat_DNatConfig) {
	if tag == "" {
		handler.log.Errorf("Cannot process DNAT config without tag")
		return
	}
	label := handler.getDnatLabel(tag)

	// Look for DNAT config using tag
	var dNat *nat.Nat44DNat_DNatConfig
	for _, dNatCfg := range *dNatCfgs {
		if dNatCfg.Label == label {
			dNat = dNatCfg
		}
	}

	// Create new DNAT config if does not exist yet
	if dNat == nil {
		dNat = &nat.Nat44DNat_DNatConfig{
			Label:      label,
			StMappings: make([]*nat.Nat44DNat_DNatConfig_StaticMapping, 0),
			IdMappings: make([]*nat.Nat44DNat_DNatConfig_IdentityMapping, 0),
		}
		*dNatCfgs = append(*dNatCfgs, dNat)
		handler.log.Debugf("Created new DNAT configuration %s", label)
	}

	// Add data to config
	switch mapping := data.(type) {
	case *nat.Nat44DNat_DNatConfig_StaticMapping:
		handler.log.Debugf("Static mapping added to DNAT %s", label)
		dNat.StMappings = append(dNat.StMappings, mapping)
	case *nat.Nat44DNat_DNatConfig_IdentityMapping:
		handler.log.Debugf("Identity mapping added to DNAT %s", label)
		dNat.IdMappings = append(dNat.IdMappings, mapping)
	}
}

// returns NAT numeric representation of provided protocol value
func (handler *natVppHandler) getNatProtocol(protocol uint8) (proto nat.Protocol) {
	switch protocol {
	case TCP:
		return nat.Protocol_TCP
	case UDP:
		return nat.Protocol_UDP
	case ICMP:
		return nat.Protocol_ICMP
	default:
		handler.log.Warnf("Unknown protocol %v", protocol)
		return 0
	}
}

func (handler *natVppHandler) getTwiceNatMode(twiceNat, selfTwiceNat uint8) nat.TwiceNatMode {
	if twiceNat > 0 {
		if selfTwiceNat > 0 {
			handler.log.Warnf("Both TwiceNAT and self-TwiceNAT are enabled")
			return 0
		}
		return nat.TwiceNatMode_ENABLED
	}
	if selfTwiceNat > 0 {
		return nat.TwiceNatMode_SELF
	}
	return nat.TwiceNatMode_DISABLED
}

func uintToBool(value uint8) bool {
	if value == 0 {
		return false
	}
	return true
}

// Obtain DNAT label from provided tag
func (handler *natVppHandler) getDnatLabel(tag string) (label string) {
	parts := strings.Split(tag, "|")
	// Tag should be in format label|mappingType|index
	if len(parts) == 0 {
		handler.log.Errorf("Unable to obtain DNAT label, incorrect mapping tag format: '%s'", tag)
		return
	}
	if len(parts) != 3 {
		handler.log.Warnf("Mapping tag has unexpected format: %s. Resolved DNAT label may not be correct", tag)
	}
	return parts[0]
}