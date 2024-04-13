// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/managedgroups"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceManagedGroup() *schema.Resource {
	return &schema.Resource{
		Description: "The boundary_managed_group data source allows you to find a Boundary managed group.",
		ReadContext: dataSourceManagedGroupRead,

		Schema: map[string]*schema.Schema{
			NameKey: {
				Description:  "The name of the managed group to retrieve.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			AuthMethodIdKey: {
				Description:  "The ID of the auth method for this managed group.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			IDKey: {
				Description: "The ID of the retrieved managed group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			DescriptionKey: {
				Description: "The description of the retrieved managed group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			ManagedGroupFilterKey: {
				Description: "The boolean expression defining a filter run against the provided information.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			GroupMemberIdsKey: {
				Description: "User IDs for members of this managed group.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
		},
	}
}

func dataSourceManagedGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	md := meta.(*metaData)

	name := d.Get(NameKey).(string)
	authMethodId := d.Get(AuthMethodIdKey).(string)

	managedGroupsClient := managedgroups.NewClient(md.client)
	managedGroupsList, err := managedGroupsClient.List(ctx, authMethodId,
		managedgroups.WithFilter(FilterWithItemNameMatches(name)),
	)
	if err != nil {
		return diag.Errorf("error calling list managed group: %v", err)
	}
	managedGroups := managedGroupsList.GetItems()
	if managedGroups == nil {
		return diag.Errorf("no managed groups found")
	}
	if len(managedGroups) == 0 {
		return diag.Errorf("no matching managed group found")
	}
	if len(managedGroups) > 1 {
		return diag.Errorf("error found more than 1 managed group")
	}

	managedGroup, err := managedGroupsClient.Read(ctx, managedGroups[0].Id)
	if err != nil {
		if apiErr := api.AsServerError(err); apiErr != nil && apiErr.Response().StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error calling read managed group: %v", err)
	}
	if managedGroup == nil {
		return diag.Errorf("managed group nil after read")
	}

	if err := setFromManagedGroupRead(d, *managedGroup.Item); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setFromManagedGroupRead(d *schema.ResourceData, managedGroup managedgroups.ManagedGroup) error {
	if err := d.Set(NameKey, managedGroup.Name); err != nil {
		return err
	}
	if err := d.Set(DescriptionKey, managedGroup.Description); err != nil {
		return err
	}
	if v, ok := managedGroup.Attributes[ManagedGroupFilterKey]; ok {
		if err := d.Set(ManagedGroupFilterKey, v); err != nil {
			return err
		}
	}
	if err := d.Set(GroupMemberIdsKey, managedGroup.MemberIds); err != nil {
		return err
	}

	d.SetId(managedGroup.Id)
	return nil
}
