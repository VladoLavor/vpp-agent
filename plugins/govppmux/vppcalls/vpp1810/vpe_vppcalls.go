//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1810

import (
	"bytes"
	"fmt"
	"strings"

	govppapi "git.fd.io/govpp.git/api"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/memclnt"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1810/vpe"
)

func init() {
	vppcalls.Versions["vpp1810"] = vppcalls.HandlerVersion{
		Msgs: append(vpe.Messages, memclnt.Messages...),
		New: func(ch govppapi.Channel) vppcalls.VpeVppAPI {
			return &VpeHandler{ch}
		},
	}
}

type VpeHandler struct {
	ch govppapi.Channel
}

func (h *VpeHandler) GetVersionInfo() (*vppcalls.VersionInfo, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	} else if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	info := &vppcalls.VersionInfo{
		Program:        string(cleanBytes(reply.Program)),
		Version:        string(cleanBytes(reply.Version)),
		BuildDate:      string(cleanBytes(reply.BuildDate)),
		BuildDirectory: string(cleanBytes(reply.BuildDirectory)),
	}

	return info, nil
}

// GetVpeInfo retrieves vpe information.
func (h *VpeHandler) GetVpeInfo() (*vppcalls.VpeInfo, error) {
	req := &vpe.ControlPing{}
	reply := &vpe.ControlPingReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	info := &vppcalls.VpeInfo{
		PID:       reply.VpePID,
		ClientIdx: reply.ClientIndex,
	}

	{
		req := &memclnt.APIVersions{}
		reply := &memclnt.APIVersionsReply{}

		if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
			return nil, err
		}

		for _, v := range reply.APIVersions {
			name := string(cleanBytes(v.Name))
			name = strings.TrimSuffix(name, ".api")
			info.ModuleVersions = append(info.ModuleVersions, vppcalls.ModuleVersion{
				Name:  name,
				Major: v.Major,
				Minor: v.Minor,
				Patch: v.Patch,
			})
		}
	}

	return info, nil
}

func (h *VpeHandler) RunCli(cmd string) (string, error) {
	req := &vpe.CliInband{
		Cmd:    []byte(cmd),
		Length: uint32(len(cmd)),
	}
	reply := &vpe.CliInbandReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	} else if reply.Retval != 0 {
		return "", fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return string(cleanBytes(reply.Reply)), nil
}

func cleanBytes(b []byte) []byte {
	return bytes.SplitN(b, []byte{0x00}, 2)[0]
}