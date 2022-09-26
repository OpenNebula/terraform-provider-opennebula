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
data "opennebula_user" "example" {
  name = "John"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula user to retrieve information for.
* `primary_group` - (Optional) Primary group ID of the user.
* `groups` - (Optional) List of secondary groups ID of the user.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the user.
* `name` - Name of the user.
* `primary_group` - Primary group ID of the user.
* `groups` - List of secondary groups ID of the user.
* `tags` - Tags of the user (Key = Value).
