// Copyright (c) 2018 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package vppcalls_test

import (
	"fmt"
	"net"
	"testing"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/bin_api/sr"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/srv6"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/srplugin"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/srplugin/vppcalls"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	. "github.com/onsi/gomega"
)

const (
	ifaceA                  = "A"
	ifaceBOutOfidxs         = "B"
	swIndexA         uint32 = 1
	invalidIPAddress        = "XYZ"
)

var (
	sidA        = sid("A::")
	sidB        = sid("B::")
	sidC        = sid("C::")
	nextHop     = net.ParseIP("B::").To16()
	nextHopIPv4 = net.ParseIP("1.2.3.4")
)

var swIfIndex = ifaceidx.NewSwIfIndex(nametoidx.NewNameToIdx(logrus.DefaultLogger(), "test-dummy-plugin",
	"sw_if_indexes", ifaceidx.IndexMetadata))

func init() {
	swIfIndex.RegisterName(ifaceA, swIndexA, nil)
}

// TestAddLocalSID tests all cases for method AddLocalSID
func TestAddLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		FailInVPP     bool
		ExpectFailure bool
		Input         *srv6.LocalSID
		Expected      *sr.SrLocalsidAddDel
	}{
		{
			Name: "addition with end behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				BaseEndFunction: &srv6.LocalSID_End{
					Psp: true,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorEnd,
				FibTable:     10,
				EndPsp:       1,
			},
		},
		{
			Name: "addition with endX behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionX: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorX,
				FibTable:     10,
				EndPsp:       1,
				SwIfIndex:    swIndexA,
				NhAddr:       nextHop,
			},
		},
		{
			Name: "addition with endT behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionT: &srv6.LocalSID_EndT{
					Psp: true,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorT,
				FibTable:     10,
				EndPsp:       1,
			},
		},
		{
			Name: "addition with endDX2 behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorDX2,
				FibTable:     10,
				EndPsp:       0,
				VlanIndex:    1,
				SwIfIndex:    swIndexA,
				NhAddr:       nextHop,
			},
		},
		{
			Name: "addition with endDX4 behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorDX4,
				FibTable:     10,
				EndPsp:       0,
				SwIfIndex:    swIndexA,
				NhAddr:       nextHopIPv4,
			},
		},
		{
			Name: "addition with endDX6 behaviour",
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX6: &srv6.LocalSID_EndDX6{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorDX6,
				FibTable:     10,
				EndPsp:       0,
				SwIfIndex:    swIndexA,
				NhAddr:       nextHop,
			},
		},
		// endDT4 and endDT6 are not fully modelled yet -> testing only current implementation
		{
			Name: "addition with endDT4 behaviour",
			Input: &srv6.LocalSID{
				FibTableID:     10,
				EndFunctionDT4: &srv6.LocalSID_EndDT4{},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorDT4,
				FibTable:     10,
				EndPsp:       0,
			},
		},
		{
			Name: "addition with endDT6 behaviour",
			Input: &srv6.LocalSID{
				FibTableID:     10,
				EndFunctionDT6: &srv6.LocalSID_EndDT6{},
			},
			Expected: &sr.SrLocalsidAddDel{
				IsDel:        0,
				LocalsidAddr: sidA,
				Behavior:     vppcalls.BehaviorDT6,
				FibTable:     10,
				EndPsp:       0,
			},
		},
		{
			Name:          "fail due to missing end function",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 0,
			},
		},
		{
			Name:          "failure propagation from VPP",
			FailInVPP:     true,
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 0,
				BaseEndFunction: &srv6.LocalSID_End{
					Psp: true,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionX: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endX behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionX: &srv6.LocalSID_EndX{
					Psp:               true,
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX2 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX2 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX2: &srv6.LocalSID_EndDX2{
					VlanTag:           1,
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHopIPv4.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX4: &srv6.LocalSID_EndDX4{
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "rejection of IPv6 addresses (addition with endDX4 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX4: &srv6.LocalSID_EndDX4{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceA,
				},
			},
		},
		{
			Name:          "missing interface in swIndexes (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX6: &srv6.LocalSID_EndDX6{
					NextHop:           nextHop.String(),
					OutgoingInterface: ifaceBOutOfidxs,
				},
			},
		},
		{
			Name:          "invalid IP address (addition with endDX6 behaviour)",
			ExpectFailure: true,
			Input: &srv6.LocalSID{
				FibTableID: 10,
				EndFunctionDX6: &srv6.LocalSID_EndDX6{
					NextHop:           invalidIPAddress,
					OutgoingInterface: ifaceA,
				},
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// prepare reply
				if td.FailInVPP {
					ctx.MockVpp.MockReply(&sr.SrLocalsidAddDelReply{Retval: 1})
				} else {
					ctx.MockVpp.MockReply(&sr.SrLocalsidAddDelReply{})
				}
				// make the call
				err := vppcalls.NewSRv6Calls().AddLocalSid(sidA, td.Input, swIfIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)
				// verify result
				if td.ExpectFailure {
					Expect(err).Should(HaveOccurred())
				} else {
					Expect(err).ShouldNot(HaveOccurred())
					Expect(ctx.MockChannel.Msg).To(Equal(td.Expected))
				}
			}()
		})
	}
}

// TestDeleteLocalSID tests all cases for method DeleteLocalSID
func TestDeleteLocalSID(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		Fail      bool
		Sid       net.IP
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple delete of local sid",
			Sid:       sidA,
			MockReply: &sr.SrLocalsidAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrLocalsidAddDel{
					IsDel:        1,
					LocalsidAddr: sidA,
				}))
			},
		},
		{
			Name:      "failure propagation from VPP",
			Sid:       sidA,
			MockReply: &sr.SrLocalsidAddDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// data and prepare case
				calls := vppcalls.NewSRv6Calls()
				localsid := &srv6.LocalSID{
					FibTableID: 10,
					BaseEndFunction: &srv6.LocalSID_End{
						Psp: true,
					},
				}
				calls.AddLocalSid(td.Sid, localsid, swIfIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)
				ctx.MockVpp.MockReply(td.MockReply)
				// make the call and verify
				err := calls.DeleteLocalSid(td.Sid, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestSetEncapsSourceAddress tests all cases for method SetEncapsSourceAddress
func TestSetEncapsSourceAddress(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		Fail      bool
		Address   string
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple SetEncapsSourceAddress",
			Address:   nextHop.String(),
			MockReply: &sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSetEncapSource{
					EncapsSource: nextHop,
				}))
			},
		},
		{
			Name:      "invalid IP address",
			Address:   invalidIPAddress,
			MockReply: &sr.SrSetEncapSourceReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:      "failure propagation from VPP",
			Address:   nextHop.String(),
			MockReply: &sr.SrSetEncapSourceReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()

				ctx.MockVpp.MockReply(td.MockReply)
				err := vppcalls.NewSRv6Calls().SetEncapsSourceAddress(td.Address, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestAddPolicy tests all cases for method AddPolicy
func TestAddPolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		Fail          bool
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple SetAddPolicy",
			BSID:          sidA,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			MockReply:     &sr.SrPolicyAddReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyAdd{
					BsidAddr:  sidA,
					FibTable:  10,
					Type:      boolToInt(false),
					IsEncap:   boolToInt(true),
					Weight:    1,
					NSegments: 3,
					Segments:  []sr.IPv6type{{Value: to16BytesArray(sidA)}, {Value: to16BytesArray(sidB)}, {Value: to16BytesArray(sidC)}},
				}))
			},
		},
		{
			Name:   "invalid SID (not IP address) in segment list",
			BSID:   sidA,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
				Weight:   1,
				Segments: []string{sidA.String(), invalidIPAddress, sidC.String()},
			},
			MockReply: &sr.SrPolicyAddReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:          "failure propagation from VPP",
			BSID:          sidA,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			MockReply:     &sr.SrPolicyAddReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// prepare reply, make call and verify
				ctx.MockVpp.MockReply(td.MockReply)
				err := vppcalls.NewSRv6Calls().AddPolicy(td.BSID, td.Policy, td.PolicySegment, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestDeletePolicy tests all cases for method DeletePolicy
func TestDeletePolicy(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name      string
		BSID      net.IP
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name:      "simple delete of policy",
			BSID:      sidA,
			MockReply: &sr.SrPolicyDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyDel{
					BsidAddr: sidA,
				}))
			},
		},
		{
			Name:      "failure propagation from VPP",
			BSID:      sidA,
			MockReply: &sr.SrPolicyDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// data and prepare case
				calls := vppcalls.NewSRv6Calls()
				policy := policy(0, true, true)
				segment := policySegment(1, sidA, sidB, sidC)
				calls.AddPolicy(td.BSID, policy, segment, logrus.DefaultLogger(), ctx.MockChannel, nil)
				ctx.MockVpp.MockReply(td.MockReply)
				// make the call and verify
				err := calls.DeletePolicy(td.BSID, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestAddPolicySegment tests all cases for method AddPolicySegment
func TestAddPolicySegment(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple addition of policy segment",
			BSID:          sidA,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			MockReply:     &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA,
					Operation: vppcalls.AddSRList,
					FibTable:  10,
					Weight:    1,
					NSegments: 3,
					Segments:  []sr.IPv6type{{Value: to16BytesArray(sidA)}, {Value: to16BytesArray(sidB)}, {Value: to16BytesArray(sidC)}},
				}))
			},
		},
		{
			Name:   "invalid SID (not IP address) in segment list",
			BSID:   sidA,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
				Weight:   1,
				Segments: []string{sidA.String(), invalidIPAddress, sidC.String()},
			},
			MockReply: &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:          "failure propagation from VPP",
			BSID:          sidA,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			MockReply:     &sr.SrPolicyModReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// prepare reply, make call and verify
				ctx.MockVpp.MockReply(td.MockReply)
				err := vppcalls.NewSRv6Calls().AddPolicySegment(td.BSID, td.Policy, td.PolicySegment, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestDeletePolicySegment tests all cases for method DeletePolicySegment
func TestDeletePolicySegment(t *testing.T) {
	// Prepare different cases
	cases := []struct {
		Name          string
		BSID          net.IP
		Policy        *srv6.Policy
		PolicySegment *srv6.PolicySegment
		SegmentIndex  uint32
		MockReply     govppapi.Message
		Verify        func(error, govppapi.Message)
	}{
		{
			Name:          "simple deletion of policy segment",
			BSID:          sidA,
			Policy:        policy(10, false, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			SegmentIndex:  111,
			MockReply:     &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrPolicyMod{
					BsidAddr:  sidA,
					Operation: vppcalls.DeleteSRList,
					SlIndex:   111,
					FibTable:  10,
					Weight:    1,
					NSegments: 3,
					Segments:  []sr.IPv6type{{Value: to16BytesArray(sidA)}, {Value: to16BytesArray(sidB)}, {Value: to16BytesArray(sidC)}},
				}))
			},
		},
		{
			Name:   "invalid SID (not IP address) in segment list",
			BSID:   sidA,
			Policy: policy(10, false, true),
			PolicySegment: &srv6.PolicySegment{
				Weight:   1,
				Segments: []string{sidA.String(), invalidIPAddress, sidC.String()},
			},
			SegmentIndex: 111,
			MockReply:    &sr.SrPolicyModReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name:          "failure propagation from VPP",
			BSID:          sidA,
			Policy:        policy(0, true, true),
			PolicySegment: policySegment(1, sidA, sidB, sidC),
			SegmentIndex:  111,
			MockReply:     &sr.SrPolicyModReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// prepare reply, make call and verify
				ctx.MockVpp.MockReply(td.MockReply)
				err := vppcalls.NewSRv6Calls().DeletePolicySegment(td.BSID, td.Policy, td.PolicySegment, td.SegmentIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

// TestAddSteering tests all cases for method AddSteering
func TestAddSteering(t *testing.T) {
	testAddRemoveSteering(t, false)
}

// TestRemoveSteering tests all cases for method RemoveSteering
func TestRemoveSteering(t *testing.T) {
	testAddRemoveSteering(t, true)
}

func testAddRemoveSteering(t *testing.T, removal bool) {
	action := "addition"
	if removal {
		action = "removal"
	}
	// Prepare different cases
	cases := []struct {
		Name      string
		Steering  *srv6.Steering
		MockReply govppapi.Message
		Verify    func(error, govppapi.Message)
	}{
		{
			Name: action + " of IPv6 L3 steering",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableID:    10,
					PrefixAddress: "1::/64",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToInt(removal),
					BsidAddr:    sidA,
					TableID:     10,
					TrafficType: vppcalls.SteerTypeIPv6,
					PrefixAddr:  net.ParseIP("1::").To16(),
					MaskWidth:   64,
				}))
			},
		},
		{
			Name: action + " of IPv4 L3 steering",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableID:    10,
					PrefixAddress: "1.2.3.4/24",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToInt(removal),
					BsidAddr:    sidA,
					TableID:     10,
					TrafficType: vppcalls.SteerTypeIPv4,
					PrefixAddr:  net.ParseIP("1.2.3.4").To16(),
					MaskWidth:   24,
				}))
			},
		},
		{
			Name: action + " of L2 steering",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L2Traffic: &srv6.Steering_L2Traffic{
					InterfaceName: ifaceA,
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(catchedMsg).To(Equal(&sr.SrSteeringAddDel{
					IsDel:       boolToInt(removal),
					BsidAddr:    sidA,
					TrafficType: vppcalls.SteerTypeL2,
					SwIfIndex:   swIndexA,
				}))
			},
		},
		{
			Name: "invalid prefix (" + action + " of IPv4 L3 steering)",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableID:    10,
					PrefixAddress: invalidIPAddress,
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "interface without index (" + action + " of L2 steering)",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L2Traffic: &srv6.Steering_L2Traffic{
					InterfaceName: ifaceBOutOfidxs,
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "invalid BSID (not IP address) as policy reference",
			Steering: &srv6.Steering{
				PolicyBSID: invalidIPAddress,
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableID:    10,
					PrefixAddress: "1::/64",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{},
			Verify: func(err error, catchedMsg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
		{
			Name: "failure propagation from VPP",
			Steering: &srv6.Steering{
				PolicyBSID: sidA.String(),
				L3Traffic: &srv6.Steering_L3Traffic{
					FibTableID:    10,
					PrefixAddress: "1::/64",
				},
			},
			MockReply: &sr.SrSteeringAddDelReply{Retval: 1},
			Verify: func(err error, msg govppapi.Message) {
				Expect(err).Should(HaveOccurred())
			},
		},
	}

	// Run all cases
	for _, td := range cases {
		t.Run(td.Name, func(t *testing.T) {
			func() { // wrapping in another function to properly teardown things inside deferred function in case of assertion failure (i.e. connection)
				ctx := vppcallmock.SetupTestCtx(t)
				defer ctx.TeardownTestCtx()
				// prepare reply, make call and verify
				ctx.MockVpp.MockReply(td.MockReply)
				var err error
				if removal {
					err = vppcalls.NewSRv6Calls().RemoveSteering(td.Steering, swIfIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)
				} else {
					err = vppcalls.NewSRv6Calls().AddSteering(td.Steering, swIfIndex, logrus.DefaultLogger(), ctx.MockChannel, nil)
				}
				td.Verify(err, ctx.MockChannel.Msg)
			}()
		})
	}
}

func sid(str string) srv6.SID {
	bsid, err := srplugin.ParseIPv6(str)
	if err != nil {
		panic(fmt.Sprintf("can't parse %q into SRv6 BSID (IPv6 address)", str))
	}
	return bsid
}

func policy(fibtableID uint32, sprayBehaviour bool, srhEncapsulation bool) *srv6.Policy {
	return &srv6.Policy{
		FibTableID:       fibtableID,
		SprayBehaviour:   sprayBehaviour,
		SrhEncapsulation: srhEncapsulation,
	}
}

func policySegment(weight uint32, sids ...srv6.SID) *srv6.PolicySegment {
	segments := make([]string, len(sids))
	for i, sid := range sids {
		segments[i] = sid.String()
	}

	return &srv6.PolicySegment{
		Weight:   weight,
		Segments: segments,
	}
}

func boolToInt(input bool) uint8 {
	if input {
		return uint8(1)
	}
	return uint8(0)
}

func to16BytesArray(slice []byte) [16]byte {
	var array [16]byte
	copy(array[:], slice)
	return array
}
