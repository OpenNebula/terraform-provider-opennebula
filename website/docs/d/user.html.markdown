---
layout: "opennebula"
page_title: "OpenNebula: opennebula_user"
sidebar_current: "docs-opennebula-datasource-user"
description: |-
  Get the user information for a given name.
---

# opennebula_user

Use this data source to retrieve the user information for a given name.

## Example Usage

```hcl
data "opennebula_user" "user" {
  name = "User"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula user to retrieve information for.

## Attribute Reference

The following attribute is exported:
* `id` - ID of the user.
* `name` - The name of the user.
* `password` - Password of the user (if set)
* `auth_driver` - Authentication Driver for User management
* `primary_group` - Primary group ID of the User.
* `groups` - List of secondary groups ID of the user.
* `quotas` - User's quotas
