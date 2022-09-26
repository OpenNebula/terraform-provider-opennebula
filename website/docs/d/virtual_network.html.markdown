---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_network"
sidebar_current: "docs-opennebula-datasource-virtual-network"
description: |-
  Get the virtual network information for a given name.
---

# opennebula_virtual_network

Use this data source to retrieve the virtual network information for a given name.

## Example Usage

```hcl
data "opennebula_virtual_network" "example" {
  name = "My_VNet"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula virtual network to retrieve information for.
* `tags` - (Optional) Virtual network tags (Key = Value).

## Attribute Reference

The following attributes are exported:

* `id` - ID of the virtual network.
* `name` - Name of the virtual network.
* `mtu` - MTU of the virtual network.
* `tags` - Tags of the virtual network (Key = Value).
