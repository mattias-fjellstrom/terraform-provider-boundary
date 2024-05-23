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
	testManagedGroupFilter = "\"12345\" in \"/token/groups\""
)

var managedGroupReadGlobal = fmt.Sprintf(`
resource "boundary_auth_method_oidc" "foo" {
	name                 = "test"
	scope_id             = "global"
	is_primary_for_scope = true
    issuer               = "https://test-update.com"
    client_id            = "foo_id_update"
    client_secret        = "foo_secret_update"
	signing_algorithms   = ["ES256"]
    api_url_prefix       = "http://localhost:9200"
	claims_scopes        = ["profile"]
}

resource "boundary_managed_group" "group" {
	name 	       = "%s"
	description    = "test"
	auth_method_id = boundary_auth_method_oidc.foo.id
	filter         = "\"12345\" in \"/token/groups\""
}

data "boundary_managed_group" "group" {
	depends_on     = [ boundary_managed_group.group ]
	name           = "%s"
	auth_method_id = boundary_auth_method_oidc.foo.id
}`, testGroupName, testGroupName)

func TestAccManagedGroupRead(t *testing.T) {
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
					resource.TestCheckResourceAttrSet("data.boundary_managed_group.group", AuthMethodIdKey),
					resource.TestCheckResourceAttr("data.boundary_managed_group.group", NameKey, testGroupName),
					resource.TestCheckResourceAttrSet("data.boundary_group.group", DescriptionKey),
					resource.TestCheckResourceAttr("data.boundary_group.group", ManagedGroupFilterKey, "\"12345\" in \"/token/groups\""),
				),
			},
		},
	})
}
