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
  id                 = "0"
}
```
## Argument Reference

* `virtual_network_id` - (Required) ID of the virtual network.
* `id` - (Required) ID of the address range.

## Attribute Reference

The following attributes are exported:

* ar_type - Type of the Address Range: IP4, IP6, IP4_6. Default is 'IP4'.
* ip4 - Start IPv4 of the range to be allocated.
* ip4_end - End IPv4 of the range to be allocated.
* ip6 - Start IPv6 of the range to be allocated.
* ip6_end - End IPv6 of the range to be allocated.
* ip6_global - Global IPv6 of the range to be allocated.
* ip6_global_end - End Global IPv6 of the range to be allocated.
* ip6_ula - ULA IPv6 of the range to be allocated.
* ip6_ula_end - End ULA IPv6 of the range to be allocated.
* size - Count of addresses in the IP range.
* mac - Start MAC of the range to be allocated.
* mac_end - End MAC of the range to be allocated.
* global_prefix - Global prefix for IP6 or IP4_6.
* ula_prefix - ULA prefix for IP6 or IP4_6.
* held_ips - List of IPs held in this address range.
* custom - Custom attributes for the address range.
