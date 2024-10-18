---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_network_address_ranges"
sidebar_current: "docs-opennebula-datasource-virtual-network-address-ranges"
description: |-
  Retrieve all address range information for a virtual network in OpenNebula.
---

# opennebula_virtual_network_address_ranges

Use this data source to retrieve information about all address ranges for a virtual network in OpenNebula.

## Example Usage

```hcl
data "opennebula_virtual_network_address_ranges" "example" {
  virtual_network_id = 123
}
```
## Argument Reference

* `virtual_network_id` - (Required) ID of the virtual network.

## Attribute Reference

The following attributes are exported:

* `address_ranges` - A list of address ranges for the specified virtual network, each containing the following attributes:
  * `id` - The ID of the address range.
  * `ar_type` - Type of the address range: IP4, IP6, etc.
  * `ip4` - Start IPv4 of the range to be allocated.
  * `size` - Count of addresses in the IP range.
