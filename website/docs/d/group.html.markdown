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

