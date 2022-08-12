---
layout: "opennebula"
page_title: "OpenNebula: opennebula_nic"
sidebar_current: "docs-opennebula-resource-nic"
description: |-
  Provides an OpenNebula virtual machine NIC resource.
---

# opennebula_nic

Provides an OpenNebula virtual machine NIC resource.

This resource allows you to manage virtual machine NICs. When applied,
a new NIC is attached to the virtual machine. When destroyed, this NIC is detached from the virtual machine.

## Example Usage

```hcl
resource "opennebula_nic" "example1" {
    vm_id           = opennebula_virtual_machine.example.id
    network_id      = opennebula_virtual_network.example.id
    ip              = "172.16.100.103"
}

resource "opennebula_nic" "example2" {
    vm_id           = opennebula_virtual_machine.example.id
    network_id      = opennebula_virtual_network.example.id
    ip              = "172.16.100.104"
}

resource "opennebula_virtual_network" "example" {
  name            = "virtual-network"
  permissions     = "660"
  bridge          = "br0"
  physical_device = "eth0"
  type            = "fw"
  mtu             = 1500
  dns             = "172.16.100.1"
  gateway         = "172.16.100.1"
  security_groups = [0]
  clusters        = [0]

  ar {
    ar_type = "IP4"
    size    = 16
    ip4     = "172.16.100.101"
  }

  tags = {
    environment = "example"
  }
}

resource "opennebula_virtual_machine" "example" {
  name        = "virtual-machine"
  description = "VM"
  cpu         = 1
  vcpu        = 1
  memory      = 1024
  permissions = "660"

  context = {
    NETWORK      = "YES"
    HOSTNAME     = "$NAME"
    START_SCRIPT = "yum upgrade"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "fr"
  }

  os {
    arch = "x86_64"
    boot = ""
  }

  tags = {
    environment = "example"
  }

  nic  {
    network_id      = opennebula_virtual_network.example.id
    ip              = "172.16.100.102"
  }

  lifecycle {
		ignore_changes = [
		  nic,
		]
	}

  hard_shutdown = true
}
```

## Argument Reference

`nic` supports the following arguments

* `vm_id` - (Required) ID of the virtual machine
* `network_id` - (Required) ID of the virtual network
* `ip` - (Optional) IP of the virtual machine on this network.
* `mac` - (Optional) MAC of the virtual machine on this network.
* `model` - (Optional) Nic model driver. Example: `virtio`.
* `virtio_queues` - (Optional) Virtio multi-queue size. Only if `model` is `virtio`.
* `physical_device` - (Optional) Physical device hosting the virtual network.
* `security_groups` - (Optional) List of security group IDs to use on the virtual network.

## Attribute Reference

* `network` - network name

## Import

It's not possible to import a `opennebula_nic`, but it's possible to migrate a `nic` section from a `opennebula_virtual_machine` resource to an `opennebula_nic`.
To do so, move all attributes from the `nic` section to the `opennebula_nic` resource and then apply.
The nic data may still live from some time in the tfstate.
