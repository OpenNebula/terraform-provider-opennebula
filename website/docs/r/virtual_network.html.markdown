---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_network"
sidebar_current: "docs-opennebula-resource-virtual-network"
description: |-
  Provides an OpenNebula virtual network resource.
---

# opennebula_virtual_network

Provides an OpenNebula virtual network resource.

This resource allows you to manage virtual networks on your OpenNebula clusters. When applied,
a new virtual network is created. When destroyed, that virtual network is removed.

## Example Usage

### Reservation of a virtual network

Allocate a new virtual network from the parent virtual network "394":

```hcl
resource "opennebula_virtual_network" "example" {
  name             = "virtual-network"
  description      = "Terraform vnet"
  reservation_vnet = 394
  reservation_size = 5
  reservation_ar_id = 1
  reservation_first_ip = "172.16.100.105"
  security_groups  = [0]
}
```

### Virtual network creation

```hcl
resource "opennebula_virtual_network" "example" {
  name            = "virtual-network"
  permissions     = "660"
  group           = opennebula_group.example.name
  bridge          = "br0"
  physical_device = "eth0"
  type            = "fw"
  mtu             = 1500
  dns             = "172.16.100.1"
  gateway         = "172.16.100.1"
  security_groups = [0]
  clusters_ids    = [0]

  ar {
    ar_type = "IP4"
    size    = 16
    ip4     = "172.16.100.101"
  }

  tags = {
    environment = "example"
  }

  template_section {
   name = "example"
   elements = {
      key1 = "value1"
   }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual network.
* `description` - (Optional) Description of the virtual network.
* `permissions` - (Optional) Permissions applied on virtual network. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `reservation_vnet` - (Optional) ID of the parent virtual network to reserve from. Conflicts with all parameters except `name`, `description`, `permissions`, `security_groups`, `group`, `reservation_ar_id`, `reservation_first_ip` and `reservation_size`.
* `reservation_size` - (Optional) Size (in address) reserved. Conflicts with all parameters except `name`, `description`, `permissions`, `security_groups`, `group`, `reservation_ar_id`, `reservation_first_ip` and `reservation_vnet`.
* `reservation_ar_id` - (Optional) ID of the address range from which to reserve the addresses. Conflicts with all parameters except `name`, `description`, `permissions`, `security_groups`, `group`, `reservation_size`, `reservation_first_ip` and `reservation_vnet`.
* `reservation_first_ip` - (Optional) The first IPv4 address to start the reservation range. Conflicts with all parameters except `name`, `description`, `permissions`, `security_groups`, `group`, `reservation_ar_id`, `reservation_size` and `reservation_vnet`.
* `security_groups` - (Optional) List of security group IDs to apply on the virtual network.
* `bridge` - (Optional) Name of the bridge interface to which the virtual network should be associated. Conflicts with `reservation_vnet` and `reservation_size`.
* `physical_device` - (Optional) Name of the physical device interface to which the virtual network should be associated. Conflicts with `reservation_vnet` and `reservation_size`.
* `type` - (Optional) Virtual network type. One of these: `dummy`, `bridge`'`fw`, `ebtables`, `802.1Q`, `vxlan` or `ovswitch`. Defaults to `bridge`. Conflicts with `reservation_vnet` and `reservation_size`.
* `clusters` - (Deprecated) List of cluster IDs where the virtual network can be use. Conflicts with `reservation_vnet` and `reservation_size`.
* `cluster_ids` - (Optional) List of cluster IDs where the virtual network can be use. Conflicts with `reservation_vnet` and `reservation_size`. Minimum 1 item.
* `vlan_id` - (Optional) ID of VLAN. Only if `type` is `802.1Q`, `vxlan` or `ovswitch`. Conflicts with `reservation_vnet`, `reservation_size` and `automatic_vlan_id`.
* `automatic_vlan_id` - (Optional) Flag to let OpenNebula scheduler to attribute the VLAN ID. Conflicts with `reservation_vnet`, `reservation_size` and `vlan_id`.
* `mtu` - (Optional) Virtual network MTU. Defaults to `1500`. Conflicts with `reservation_vnet` and `reservation_size`.
* `guest_mtu` - (Optional) MTU of the network caord on the virtual machine. **Cannot be greater than `mtu`**. Defaults to `1500`. Conflicts with `reservation_vnet` and `reservation_size`.
* `gateway` - (Optional) IP of the gateway. Conflicts with `reservation_vnet` and `reservation_size`.
* `network_mask` - (Optional) Network mask. Conflicts with `reservation_vnet` and `reservation_size`.
* `dns` - (Optional) Text String containing a space separated list of DNS IPs. Conflicts with `reservation_vnet` and `reservation_size`.
* `network_address` - (Optional) Base network address. Conflicts with `reservation_vnet` and `reservation_size`.
* `search_domain` - (Optional) Default search domains for DNS resolution. Conflicts with `reservation_vnet` and `reservation_size`.
* `ar` - (Deprecated) List of address ranges. See [Address Range Parameters](#address-range-parameters) below for more details. Conflicts with `reservation_vnet` and `reservation_size`.
* `hold_ips` - (Deprecated) Hold Ips from any Address Range of the Virtual Network. The IP must be available to be held`. Conflicts with`reservation_vnet` and `reservation_size`.
* `group` - (Optional) Name of the group which owns the virtual network. Defaults to the caller primary group.
* `tags` - (Optional) Virtual Network tags (Key = Value).
* `lock` - (Optional) Lock the virtual network with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.
* `template_section` - (Optional) Allow to add a custom vector. See [Template section parameters](#template-section-parameters)

### Address Range parameters

`ar` supports the following arguments:

* `ar_type` - (Optional) Address range type. Supported values: `IP4`, `IP6`, `IP6_STATIC`, `IP4_6` or `IP4_6_STATIC` or `ETHER`. Defaults to `IP4`.
* `ip4` - (Optional) Starting IPv4 address of the range. Required if `ar_type` is `IP4` or `IP4_6`.
* `ip6` - (Optional) Starting IPv6 address of the range. Required if `ar_type` is `IP6_STATIC` or `IP4_6_STATIC`.
* `size` - (Required) Address range size.
* `mac` - (Optional) Starting MAC Address of the range.
* `global_prefix` - (Optional) Global prefix for `IP6` or `IP_4_6`.
* `ula_prefix` - (Optional) ULA prefix for `IP6` or `IP_4_6`.
* `prefix_length` - (Optional) Prefix length. Only needed for `IP6_STATIC` or `IP4_6_STATIC`

### Template section parameters

`template_section` supports the following arguments:

* `name` - (Optional) The vector name.
* `elements` - (Optional) Collection of custom tags.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the virtual network.
* `uid` - User ID whom owns the virtual network.
* `gid` - Group ID which owns the virtual network.
* `uname` - User Name whom owns the virtual network.
* `gname` - Group Name which owns the virtual network.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

### Address range computed attributes

* `id` - ID of the address range
* `computed_ip6` - Starting IPv6 address of the range.
* `computed_mac` - Starting MAC Address of the range.
* `computed_global_prefix` - Global prefix for type `IP6` or `IP_4_6`.
* `computed_ula_prefix` - ULA prefix for type `IP6` or `IP_4_6`.

## Import

`opennebula_virtual_network` can be imported using its ID:

```shell
terraform import opennebula_virtual_network.example 123
```
