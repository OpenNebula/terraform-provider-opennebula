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
    * `ar_type` - Type of the Address Range: `IP4`, `IP6`, `IP4_6`. Default is `IP4`.
    * `ip4` - Start IPv4 of the allocated range.
    * `ip4_end` - End IPv4 of the allocated range.
    * `ip6` - Start IPv6 of the allocated range.
    * `ip6_end` - End IPv6 of the allocated range.
    * `ip6_global` - Global IPv6 of the allocated range.
    * `ip6_global_end` - End Global IPv6 of the allocated range.
    * `ip6_ula` - ULA IPv6 of the allocated range.
    * `ip6_ula_end` - End ULA IPv6 of the allocated range.
    * `size` - Count of addresses in the IP range.
    * `mac` - Start MAC of the allocated range.
    * `mac_end` - End MAC of the allocated range.
    * `global_prefix` - Global prefix for `IP6` or `IP4_6`.
    * `ula_prefix` - ULA prefix for `IP6` or `IP4_6`.
    * `held_ips` - List of IPs held in this address range.
    * `custom` - Custom attributes for the address range.
