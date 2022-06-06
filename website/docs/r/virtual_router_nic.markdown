---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_router_nic"
sidebar_current: "docs-opennebula-resource-virtual-router"
description: |-
  Provides an OpenNebula virtual router resource.
---

# opennebula_virtual_router_nic

Provides an OpenNebula virtual router resource.

## Example Usage

```hcl
resource "opennebula_virtual_router_instance_template" "example" {
  name        = "virtual-router-instance-template"
  permissions = "642"
  group       = "oneadmin"
  cpu         = "0.5"
  vcpu        = "1"
  memory      = "512"

  context = {
    dns_hostname = "yes"
    network      = "YES"
  }

  graphics {
    keymap = "en-us"
    listen = "0.0.0.0"
    type   = "VNC"
  }

  os {
    arch = "x86_64"
    boot = ""
  }

  tags = {
    env = "prod"
  }
}

resource "opennebula_virtual_router" "example" {
  name        = "virtual-router"
  permissions = "642"
  group       = "oneadmin"
  description = "This is an example of virtual router"

  instance_template_id = opennebula_virtual_router_instance_template.example.id

  lock = "USE"

  tags = {
    environment = "example"
  }
}

resource "opennebula_virtual_router_instance" "example" {
  name        = "virtual-router-instance"
  group       = "oneadmin"
  permissions = "642"
  memory      = 128
  cpu         = 0.1

  virtual_router_id = opennebula_virtual_router.example.id

  tags = {
    environment = "example"
  }
}

resource "opennebula_virtual_network" "example" {
  name   = "virtual-network"
  type   = "bridge"
  bridge = "onebr"
  mtu    = 1500

  ar {
    ar_type = "IP4"
    size    = 12
    ip4     = "172.16.100.130"
  }

  permissions     = "642"
  group           = "oneadmin"
  security_groups = [0]
  clusters        = [0]
}

resource "opennebula_virtual_router_nic" "example" {
  virtual_router_id = opennebula_virtual_router.example.id
  network_id        = opennebula_virtual_network.example.id
}
```

## Argument Reference

The following arguments are supported:

* `virtual_router_id` - (Required) The ID of the parent virtual router resource.
* `network_id` - (Required) ID of the virtual network to attach.
* `model` - (Optional) Nic model driver. Example: `virtio`.
* `virtio_queues` - (Optional) Virtio multi-queue size. Only if `model` is `virtio`.
* `physical_device` - (Optional) Physical device hosting the virtual network.
* `security_groups` - (Optional) List of security group IDs to use on the virtual network.

## Attribute Reference

The following attribute are exported:

* `network` - Name of the virtual network to attach.
* `model` - Nic model driver. Example: `virtio`.
* `virtio_queues` - Virtio multi-queue size. Only if `model` is `virtio`.
* `physical_device` - Physical device hosting the virtual network.
* `security_groups` - List of security group IDs to use on the virtual network.

## Import

`opennebula_virtual_router_nic` can be imported using its ID:

```shell
terraform import opennebula_virtual_router_nic.example 123
```
