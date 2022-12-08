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
data "opennebula_group" "example" {
  name = "My_Service_Group"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula group to retrieve information for.
* `tags` - (Optional) Tags associated to the Image.

## Attribute Reference

The following attribute is exported:

* `id` - ID of the group.
* `name` - Name of the group.
* `admins` - List of Administrator user IDs part of the group.
* `tags` - Tags of the group (Key = Value).
