---
layout: "opennebula"
page_title: "OpenNebula: opennebula_group"
sidebar_current: "docs-opennebula-datasource-group"
description: |-
  Get the group information for a given name.
---

# opennebula_group

Use this data source to retrieve the group information for a given name.

## Example Usage

```hcl
data "opennebula_group" "ExistingGroup" {
  name = "My_Service_Group"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula group to retrieve information for.

## Attribute Reference

The following attribute is exported:
* `id` - ID of the group.
* `users` - List of User IDs part of the group.
* `admins` - List of Administrator user IDs part of the group.
* `quotas` - Quotas configured for the group.

