// +build fvtests

// Copyright (c) 2017 Tigera, Inc. All rights reserved.
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

package fv_test

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/projectcalico/felix/fv/containers"
	"github.com/projectcalico/felix/fv/workload"
	"github.com/projectcalico/libcalico-go/lib/api"
	"github.com/projectcalico/libcalico-go/lib/client"
	"github.com/projectcalico/libcalico-go/lib/numorstring"
)

// Setup for planned further FV tests:
//
//     | +-----------+ +-----------+ |  | +-----------+ +-----------+ |
//     | | service A | | service B | |  | | service C | | service D | |
//     | | 10.65.0.2 | | 10.65.0.3 | |  | | 10.65.0.4 | | 10.65.0.5 | |
//     | | port 9002 | | port 9003 | |  | | port 9004 | | port 9005 | |
//     | | np 109002 | | port 9003 | |  | | port 9004 | | port 9005 | |
//     | +-----------+ +-----------+ |  | +-----------+ +-----------+ |
//     +-----------------------------+  +-----------------------------+
			Expect(w[1]).To(HaveConnectivityTo(w[0], 32010))
			Expect(etcd).To(HaveConnectivityTo(w[1], 32011))
			Expect(etcd).To(HaveConnectivityTo(w[0], 32010))
		})

		Context("with pre-DNAT policy to prevent access from outside", func() {

			BeforeEach(func() {
				policy := api.NewPolicy()
				policy.Metadata.Name = "deny-ingress"
				order := float64(20)
				policy.Spec.Order = &order
				policy.Spec.PreDNAT = true
				policy.Spec.IngressRules = []api.Rule{{Action: "deny"}}
				policy.Spec.Selector = "has(host-endpoint)"
				_, err := client.Policies().Create(policy)
				Expect(err).NotTo(HaveOccurred())

				hostEp := api.NewHostEndpoint()
				hostEp.Metadata.Name = "felix-eth0"
				hostEp.Metadata.Node = felix.Hostname
				hostEp.Metadata.Labels = map[string]string{"host-endpoint": "true"}
				hostEp.Spec.InterfaceName = "eth0"
				_, err = client.HostEndpoints().Create(hostEp)
				Expect(err).NotTo(HaveOccurred())
			})

			It("etcd cannot connect", func() {
				Expect(w[0]).To(HaveConnectivityTo(w[1], 32011))
				Expect(w[1]).To(HaveConnectivityTo(w[0], 32010))
				Expect(etcd).NotTo(HaveConnectivityTo(w[1], 32011))
				Expect(etcd).NotTo(HaveConnectivityTo(w[0], 32010))
			})

			Context("with pre-DNAT policy to open pinhole to 32010", func() {

				BeforeEach(func() {
					policy := api.NewPolicy()
					policy.Metadata.Name = "allow-ingress-32010"
					order := float64(10)
					policy.Spec.Order = &order
					policy.Spec.PreDNAT = true
					protocol := numorstring.ProtocolFromString("tcp")
					ports, _ := numorstring.PortFromRange(1, 65535)   // pass
					ports, _ = numorstring.PortFromRange(1, 10000)    // pass
					ports, _ = numorstring.PortFromRange(8000, 8055)  // pass
					ports, _ = numorstring.PortFromRange(8056, 10000) // fail
					ports = numorstring.SinglePort(8055)
					policy.Spec.IngressRules = []api.Rule{{
						Action:   "allow",
						Protocol: &protocol,
						Destination: api.EntityRule{Ports: []numorstring.Port{
							ports,
						}},
					}}
					policy.Spec.Selector = "has(host-endpoint)"
					_, err := client.Policies().Create(policy)
					Expect(err).NotTo(HaveOccurred())
				})

				// Pending because currently fails - investigation needed.
				PIt("etcd can connect to 32010 but not 32011", func() {
					Expect(w[0]).To(HaveConnectivityTo(w[1], 32011))
					Expect(w[1]).To(HaveConnectivityTo(w[0], 32010))
					//Expect(etcd).NotTo(HaveConnectivityTo(w[1], 32011))
					var success bool
					var err error
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					success, err = HaveConnectivityTo(w[0], 32010).Match(etcd)
					Expect(success).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())
					Expect(etcd).NotTo(HaveConnectivityTo(w[0], 32010))
				})
			})
		})
	})
})
