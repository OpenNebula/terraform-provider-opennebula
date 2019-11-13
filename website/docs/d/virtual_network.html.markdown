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
data "opennebula_virtual_network" "ExistingVNet" {
  name = "My_VNet"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula virtual network to retrieve information for.

