---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_network_address_range"
sidebar_current: "docs-opennebula-datasource-virtual-network-address-range"
description: |-
  Retrieve address range information for a virtual network in OpenNebula.
---

# opennebula_virtual_network_address_range

Use this data source to retrieve address range information for a virtual network in OpenNebula.

## Example Usage

```hcl
data "opennebula_virtual_network_address_range" "example" {
  virtual_network_id = 123
}
```
## Argument Reference

* `virtual_network_id` - (Required) ID of the virtual network.

## Argument Reference

* `virtual_network_id` - (Required) ID of the virtual network.

## Attribute Reference

The following attributes are exported:

* `address_ranges` - A list of address ranges for the specified virtual network, each containing the following attributes:
  * `ar_type` - Type of the Address Range: IP4, IP6.
  * `ip4` - Start IPv4 of the range to be allocated.
  * `size` - Count of addresses in the IP range.
  * `mac` - Start MAC of the range to be allocated.
  * `global_prefix` - Global prefix for IP6 or IP4_6.
  * `held_ips` - List of IPs held in this address range.
  * `custom` - Custom attributes for the address range.
