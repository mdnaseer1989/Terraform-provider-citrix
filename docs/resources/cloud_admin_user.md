---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_cloud_admin_user Resource - citrix"
subcategory: "Citrix Cloud"
description: |-
  Manages an administrator user for cloud environment.
---

# citrix_cloud_admin_user (Resource)

Manages an administrator user for cloud environment.

## Example Usage

```terraform
resource "citrix_cloud_admin_user" "example-admin-user" {
  access_type = "Full"
  email = "example-admin@citrix.com"
  provider_type = "CitrixSts"
  type = "AdministratorUser"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_type` (String) Access Type of the user. Currently, this attribute can only be set to `Full`
- `email` (String) Email of the user where the invitation link will be sent.
- `provider_type` (String) Identity provider for the administrator or group you want to add. Currently, this attribute can only be set to `CitrixSts`.
- `type` (String) Type of administrator being added. Currently, this attribute can only be set to `AdministratorUser`.

### Optional

- `display_name` (String) Display name for the user.
- `first_name` (String) First name of the user.
- `last_name` (String) Last name of the user.

### Read-Only

- `ucoid` (String) Universal claim organization identifier of the administrator.
- `user_id` (String) Id of the administrator.

## Import

Import is supported using the following syntax:

```shell
# Admin User can be imported by specifying their email
terraform import citrix_cloud_admin_user.example-admin-user example-admin@citrix.com
```