data "boundary_scope" "organization" {
  name     = "organization"
  scope_id = "global"
}

data "boundary_auth_method" "oidc" {
  scope_id = data.boundary_scope.organization.id
  name     = "oidc_auth_method_name"
}

data "boundary_managed_group" "administrators" {
  name           = "administrators"
  auth_method_id = data.boundary_auth_method.oidc.id
}
