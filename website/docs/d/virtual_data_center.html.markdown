---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_data_center"
sidebar_current: "docs-opennebula-datasource-virtual-data-center"
description: |-
  Get the virtual data center information for a given name.
---

# opennebula_virtual_data_center

Use this data source to retrieve the virtual data center information for a given name.

## Example Usage

```hcl
data "opennebula_virtual_data_center" "ExistingVdc" {
  name = "My_VDC"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula virtual data center to retrieve information for.

