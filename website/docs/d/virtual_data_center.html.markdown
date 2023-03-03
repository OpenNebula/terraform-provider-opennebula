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
data "opennebula_virtual_data_center" "example" {
  name = "My_VDC"
}
```

## Argument Reference

* `id` - (Optional) ID of the virtual data center.
* `name` - (Optional) The OpenNebula virtual data center to retrieve information for.
* `tags` - (Optional) Virtual data center tags (Key = Value).

## Attribute Reference

* `id` - ID of the virtual data center.
* `name` - Name of the virtual data center.
* `tags` - Tags of the virtual data center (Key = Value).
