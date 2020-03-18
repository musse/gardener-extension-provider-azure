// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation_test

import (
	apisazure "github.com/gardener/gardener-extension-provider-azure/pkg/apis/azure"
	. "github.com/gardener/gardener-extension-provider-azure/pkg/apis/azure/validation"

	. "github.com/gardener/gardener/pkg/utils/validation/gomega"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("InfrastructureConfig validation", func() {
	var (
		infrastructureConfig *apisazure.InfrastructureConfig
		nodes                string
		resourceGroup        = "shoot--test--foo"

		pods        = "100.96.0.0/11"
		services    = "100.64.0.0/13"
		vnetCIDR    = "10.0.0.0/8"
		invalidCIDR = "invalid-cidr"

		fldPath *field.Path
	)

	BeforeEach(func() {
		nodes = "10.250.0.0/16"
		infrastructureConfig = &apisazure.InfrastructureConfig{
			Networks: apisazure.NetworkConfig{
				Workers: "10.250.3.0/24",
				VNet: apisazure.VNet{
					CIDR: &vnetCIDR,
				},
			},
		}
	})

	Describe("#ValidateInfrastructureConfig", func() {
		It("should forbid specifying a resource group configuration", func() {
			infrastructureConfig.ResourceGroup = &apisazure.ResourceGroup{}

			errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

			Expect(errorList).To(ConsistOfFields(Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("resourceGroup"),
			}))
		})

		Context("vnet", func() {
			It("should forbid specifying a vnet name without resource group", func() {
				vnetName := "existing-vnet"
				infrastructureConfig.Networks.VNet = apisazure.VNet{
					Name: &vnetName,
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.vnet"),
						"Detail": Equal("specifying an existing vnet name require a vnet name and vnet resource group"),
					}))
			})

			It("should forbid specifying a vnet resource group without name", func() {
				vnetGroup := "existing-vnet-rg"
				infrastructureConfig.Networks.VNet = apisazure.VNet{
					ResourceGroup: &vnetGroup,
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.vnet"),
						"Detail": Equal("specifying an existing vnet name require a vnet name and vnet resource group"),
					}))
			})

			It("should forbid specifying existing vnet plus a vnet cidr", func() {
				name := "existing-vnet"
				vnetGroup := "existing-vnet-rg"
				infrastructureConfig.Networks.VNet = apisazure.VNet{
					Name:          &name,
					ResourceGroup: &vnetGroup,
					CIDR:          &vnetCIDR,
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.vnet.cidr"),
						"Detail": Equal("specifying a cidr for an existing vnet is not possible"),
					}))
			})

			It("should forbid specifying existing vnet in same resource group", func() {
				name := "existing-vnet"
				infrastructureConfig.Networks.VNet = apisazure.VNet{
					Name:          &name,
					ResourceGroup: &resourceGroup,
				}
				infrastructureConfig.ResourceGroup = &apisazure.ResourceGroup{
					Name: resourceGroup,
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(
					Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("resourceGroup"),
					},
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.vnet.resourceGroup"),
						"Detail": Equal("the vnet resource group must not be the same as the cluster resource group"),
					}))
			})

			It("should pass if no vnet cidr is specified and default is applied", func() {
				nodes = "10.250.3.0/24"
				infrastructureConfig.ResourceGroup = nil
				infrastructureConfig.Networks = apisazure.NetworkConfig{
					Workers: "10.250.3.0/24",
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)
				Expect(errorList).To(HaveLen(0))
			})
		})

		Context("CIDR", func() {
			It("should forbid invalid VNet CIDRs", func() {
				infrastructureConfig.Networks.VNet.CIDR = &invalidCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.vnet.cidr"),
					"Detail": Equal("invalid CIDR address: invalid-cidr"),
				}))
			})

			It("should forbid invalid workers CIDR", func() {
				infrastructureConfig.Networks.Workers = invalidCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.workers"),
					"Detail": Equal("invalid CIDR address: invalid-cidr"),
				}))
			})

			It("should forbid empty workers CIDR", func() {
				infrastructureConfig.Networks.Workers = ""

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.workers"),
						"Detail": Equal("invalid CIDR address: "),
					}))
			})

			It("should forbid workers which are not in VNet and Nodes CIDR", func() {
				notOverlappingCIDR := "1.1.1.1/32"
				infrastructureConfig.Networks.Workers = notOverlappingCIDR

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.workers"),
					"Detail": Equal(`must be a subset of "" ("10.250.0.0/16")`),
				}, Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.workers"),
					"Detail": Equal(`must be a subset of "networks.vnet.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid Pod CIDR to overlap with VNet CIDR", func() {
				podCIDR := "10.0.0.1/32"

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &podCIDR, &services, fldPath)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal(""),
					"Detail": Equal(`must not be a subset of "networks.vnet.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid Services CIDR to overlap with VNet CIDR", func() {
				servicesCIDR := "10.0.0.1/32"

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &servicesCIDR, fldPath)

				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal(""),
					"Detail": Equal(`must not be a subset of "networks.vnet.cidr" ("10.0.0.0/8")`),
				}))
			})

			It("should forbid non canonical CIDRs", func() {
				vpcCIDR := "10.0.0.3/8"
				nodeCIDR := "10.250.0.3/16"
				podCIDR := "100.96.0.4/11"
				serviceCIDR := "100.64.0.5/13"
				workers := "10.250.3.8/24"

				infrastructureConfig.Networks.Workers = workers
				infrastructureConfig.Networks.VNet = apisazure.VNet{CIDR: &vpcCIDR}

				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodeCIDR, &podCIDR, &serviceCIDR, fldPath)

				Expect(errorList).To(HaveLen(2))
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.vnet.cidr"),
					"Detail": Equal("must be valid canonical CIDR"),
				}, Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.workers"),
					"Detail": Equal("must be valid canonical CIDR"),
				}))
			})
		})

		Context("Identity", func() {
			It("should return no errors for using an identity", func() {
				infrastructureConfig.Identity = &apisazure.IdentityConfig{
					Name:          "test-identiy",
					ResourceGroup: "identity-resource-group",
				}
				Expect(ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)).To(BeEmpty())
			})

			It("should return errors because no name or resource group is given", func() {
				infrastructureConfig.Identity = &apisazure.IdentityConfig{
					Name: "test-identiy",
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("identity"),
				}))
			})
		})

		Context("NatGateway", func() {
			It("should return no errors using a NatGateway for a zoned cluster", func() {
				infrastructureConfig.Zoned = true
				infrastructureConfig.Networks.NatGateway = &apisazure.NatGatewayConfig{Enabled: true}
				Expect(ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)).To(BeEmpty())
			})

			It("should return an error using a NatGateway for a non zoned cluster", func() {
				infrastructureConfig.Zoned = false
				infrastructureConfig.Networks.NatGateway = &apisazure.NatGatewayConfig{}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)
				Expect(errorList).To(HaveLen(1))
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.natGateway"),
					"Detail": Equal("NatGateway is currently only supported for zoned cluster"),
				}))
			})

			It("should return an error using a NatGateway as it not enabled but ip addresses are configured", func() {
				infrastructureConfig.Zoned = true
				infrastructureConfig.Networks.NatGateway = &apisazure.NatGatewayConfig{
					Enabled: false,
					IPAddresses: []apisazure.AzResourceReference{{
						Name:          "test-ip",
						ResourceGroup: "test-rg",
					}},
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)
				Expect(errorList).To(HaveLen(1))
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.natGateway"),
					"Detail": Equal("NatGateway is not enabled but ip addresses or ip ranges are specified"),
				}))
			})

			It("should return an error using a NatGateway as referenced ip is invalid", func() {
				infrastructureConfig.Zoned = true
				infrastructureConfig.Networks.NatGateway = &apisazure.NatGatewayConfig{
					Enabled: true,
					IPAddresses: []apisazure.AzResourceReference{{
						Name: "test-ip",
					}},
					IPAddressRanges: []apisazure.AzResourceReference{{
						ResourceGroup: "test-rg",
					}},
				}
				errorList := ValidateInfrastructureConfig(infrastructureConfig, &nodes, &pods, &services, fldPath)
				Expect(errorList).To(HaveLen(2))
				Expect(errorList).To(ConsistOfFields(Fields{
					"Type":   Equal(field.ErrorTypeInvalid),
					"Field":  Equal("networks.natGateway.ipAddresses[0]"),
					"Detail": Equal("resourceGroup must be set"),
				},
					Fields{
						"Type":   Equal(field.ErrorTypeInvalid),
						"Field":  Equal("networks.natGateway.ipAddressRanges[0]"),
						"Detail": Equal("name must be set"),
					}))
			})
		})
	})

	Describe("#ValidateInfrastructureConfigUpdate", func() {
		It("should return no errors for an unchanged config", func() {
			Expect(ValidateInfrastructureConfigUpdate(infrastructureConfig, infrastructureConfig, fldPath)).To(BeEmpty())
		})

		It("should forbid changing the resource group section", func() {
			newInfrastructureConfig := infrastructureConfig.DeepCopy()
			newInfrastructureConfig.ResourceGroup = &apisazure.ResourceGroup{}

			errorList := ValidateInfrastructureConfigUpdate(infrastructureConfig, newInfrastructureConfig, fldPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("resourceGroup"),
			}))))
		})

		It("should forbid changing the network section", func() {
			newInfrastructureConfig := infrastructureConfig.DeepCopy()
			newCIDR := "1.2.3.4/5"
			newInfrastructureConfig.Networks.VNet.CIDR = &newCIDR

			errorList := ValidateInfrastructureConfigUpdate(infrastructureConfig, newInfrastructureConfig, fldPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeInvalid),
				"Field": Equal("networks.vnet"),
			}))))
		})

		It("should forbid moving a zoned cluster to a non zoned cluster", func() {
			newInfrastructureConfig := infrastructureConfig.DeepCopy()
			infrastructureConfig.Zoned = true

			errorList := ValidateInfrastructureConfigUpdate(infrastructureConfig, newInfrastructureConfig, fldPath)

			Expect(errorList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeForbidden),
				"Field": Equal("zoned"),
			}))))
		})
	})
})
