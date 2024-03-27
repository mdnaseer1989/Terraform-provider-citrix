---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "citrix_policy_set Resource - citrix"
subcategory: ""
description: |-
  Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority. This feature will be officially supported for On-Premises with DDC version 2402 and above and will be made available for Cloud soon.
---

# citrix_policy_set (Resource)

Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority. This feature will be officially supported for On-Premises with DDC version 2402 and above and will be made available for Cloud soon.

## Example Usage

```terraform
resource "citrix_policy_set" "example-policy-set" {
    name = "example-policy-set"
    description = "This is an example policy set description"
    type = "DeliveryGroupPolicies"
    scopes = [ "All", citrix_admin_scope.example-admin-scope.name ]
    policies = [
        {
            name = "test-policy-with-priority-0"
            description = "Test policy in the example policy set with priority 0"
            is_enabled = true
            policy_settings = [
                {
                    name = "AdvanceWarningPeriod"
                    value = "13:00:00"
                    use_default = false
                },
            ]
            policy_filters = [
                {
                    type = "DesktopGroup"
                    data = jsonencode({
                        "server" = "20.185.46.142"
                        "uuid" = citrix_policy_set.example-delivery-group.id
                    })
                    is_enabled = true
                    is_allowed = true
                },
            ]
        },
        {
            name = "test-policy-with-priority-1"
            description = "Test policy in the example policy set with priority 1"
            is_enabled = false
            policy_settings = []
            policy_filters = []
        }
    ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the policy set.
- `policies` (Attributes List) Ordered list of policies. The order of policies in the list determines the priority of the policies. (see [below for nested schema](#nestedatt--policies))
- `scopes` (Set of String) The names of the scopes for the policy set to apply on.
- `type` (String) Type of the policy set. Type can be one of `SitePolicies`, `DeliveryGroupPolicies`, `SiteTemplates`, or `CustomTemplates`.

### Optional

- `description` (String) Description of the policy set.

### Read-Only

- `id` (String) GUID identifier of the policy set.
- `is_assigned` (Boolean) Indicate whether the policy set is being assigned to delivery groups.

<a id="nestedatt--policies"></a>
### Nested Schema for `policies`

Required:

- `is_enabled` (Boolean) Indicate whether the policy is being enabled.
- `name` (String) Name of the policy.
- `policy_filters` (Attributes Set) Set of policy filters. (see [below for nested schema](#nestedatt--policies--policy_filters))
- `policy_settings` (Attributes Set) Set of policy settings. (see [below for nested schema](#nestedatt--policies--policy_settings))

Optional:

- `description` (String) Description of the policy.

<a id="nestedatt--policies--policy_filters"></a>
### Nested Schema for `policies.policy_filters`

Required:

- `is_allowed` (Boolean) Indicate the filtered policy is allowed or denied if the filter condition is met.
- `is_enabled` (Boolean) Indicate whether the policy is being enabled.
- `type` (String) Type of the policy filter. Type can be one of `AccessControl`, `BranchRepeater`, `ClientIP`, `ClientName`, `DesktopGroup`, `DesktopKind`, `OU`, `User`, and `DesktopTag`

Optional:

- `data` (String) Data of the policy filter.


<a id="nestedatt--policies--policy_settings"></a>
### Nested Schema for `policies.policy_settings`

Required:

- `name` (String) Name of the policy setting name.
- `use_default` (Boolean) Indicate whether using default value for the policy setting.
- `value` (String) Value of the policy setting.

## Import

Import is supported using the following syntax:

```shell
# Policy and Policy Set Association can be imported by specifying the Policy GUID
terraform import citrix_policy_set.example 00000000-0000-0000-0000-000000000000
```