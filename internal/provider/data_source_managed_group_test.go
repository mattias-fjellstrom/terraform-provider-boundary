// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/boundary/testing/controller"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	testManagedGroupName   = "test_managed_group"
	testAuthMethodId       = "auth-method-id"
	testManagedGroupFilter = "\"12345\" in /token/groups"
)

var managedGroupReadGlobal = fmt.Sprintf(`
resource "boundary_managed_group" "group" {
	name 	       = "%s"
	description    = "test"
	auth_method_id = "%s"
	filter         = "%s"
}

data "boundary_managed_group" "group" {
	depends_on     = [ boundary_managed_group.group ]
	name           = "%s"
	auth_method_id = "%s"
}`, testGroupName, testAuthMethodId, testManagedGroupFilter, testGroupName, testAuthMethodId)

func TestAccManagedGroupReadGlobal(t *testing.T) {
	tc := controller.NewTestController(t, tcConfig...)
	defer tc.Shutdown()
	url := tc.ApiAddrs()[0]

	var provider *schema.Provider
	resource.Test(t, resource.TestCase{
		ProviderFactories: providerFactories(&provider),
		Steps: []resource.TestStep{
			{
				Config: testConfig(url, fooOrg, managedGroupReadGlobal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedGroupResourceExists(provider, "boundary_managed_group.group"),
					resource.TestCheckResourceAttrSet("data.boundary_managed_group.group", IDKey),
					resource.TestCheckResourceAttr("data.boundary_managed_group.group", AuthMethodIdKey, testAuthMethodId),
					resource.TestCheckResourceAttr("data.boundary_managed_group.group", NameKey, testGroupName),
					resource.TestCheckResourceAttrSet("data.boundary_group.group", DescriptionKey),
					resource.TestCheckResourceAttr("data.boundary_group.group", ManagedGroupFilterKey, testManagedGroupFilter),
				),
			},
		},
	})
}
