---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_network_address_range"
sidebar_current: "docs-opennebula-resource-virtual-network-address-range"
description: |-
  Provides an OpenNebula virtual network address range resource.
---

# opennebula_virtual_network_address_range

Provides an OpenNebula virtual network address range resource. When applied, a new address range is added to the virtual network. When destroyed, the address range is removed from the virtual network.

## Example Usage

```hcl
resource "opennebula_virtual_network" "test" {
  name         = "test-virtual_network"
  type         = "bridge"
  bridge       = "onebr"
  mtu          = 1500
  gateway      = "172.16.100.1"
  dns          = "172.16.100.1"
  network_mask = "255.255.255.0"

  # deprecated
  ar {
    ar_type = "IP4"
    size    = 15
    ip4     = "172.16.100.170"
  }

  permissions     = "642"
  group           = "oneadmin"
  security_groups = [0]
  clusters        = [0]
  tags = {
    env      = "prod"
    customer = "test"
  }

  lifecycle {
    ignore_changes = [ar]
  }
}

resource "opennebula_virtual_network_address_range" "test" {
  virtual_network_id = opennebula_virtual_network.test.id
  ar_type            = "IP4"
  mac                = "02:00:ac:10:64:6e"
  size               = 15
  ip4                = "172.16.100.110"

  hold_ips = ["172.16.100.112", "172.16.100.114"]

  lifecycle {
    ignore_changes = [hold_ips]
  }
}
```

## Argument Reference

The following arguments are supported:

* `virtual_network_id` - (Required) ID of the virtual network
* `ar_type` - (Optional) Address range type. Supported values: `IP4`, `IP6`, `IP6_STATIC`, `IP4_6` or `IP4_6_STATIC` or `ETHER`. Defaults to `IP4`.
* `ip4` - (Optional) Starting IPv4 address of the range. Required if `ar_type` is `IP4` or `IP4_6`.
* `ip6` - (Optional) Starting IPv6 address of the range. Required if `ar_type` is `IP6_STATIC` or `IP4_6_STATIC`.
* `size` - (Required) Address range size.
* `mac` - (Optional) Starting MAC Address of the range.
* `global_prefix` - (Optional) Global prefix for `IP6` or `IP_4_6`.
* `ula_prefix` - (Optional) ULA prefix for `IP6` or `IP_4_6`.
* `prefix_length` - (Optional) Prefix length. Only needed for `IP6_STATIC` or `IP4_6_STATIC`

## Attribute Reference

The following attribute are exported:

* `mac` - Starting MAC Address of the range.

## Import


`opennebula_virtual_network_address_range` can be imported using a composed ID:

```sh
terraform import opennebula_virtual_network_address_range.example vnet_id:ar_id
```
