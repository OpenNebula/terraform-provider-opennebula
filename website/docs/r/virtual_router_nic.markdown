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
resource "opennebula_virtual_router_instance_template" "test" {
  name        = "testacc-vr-template"
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

resource "opennebula_virtual_router" "vrouter" {
  name        = "testacc-vr"
  permissions = "642"
  group       = "oneadmin"
  description = "This is an example of virtual router"

  instance_template_id = opennebula_virtual_router_instance_template.test.id

  lock = "USE"
  tags = {
    environment = "test"
  }
}

resource "opennebula_virtual_router_instance" "test" {
  name        = "testacc-vr-virtual-machine"
  group       = "oneadmin"
  permissions = "642"
  memory      = 128
  cpu         = 0.1

  virtual_router_id = opennebula_virtual_router.test.id

  tags = {
    customer = "1"
  }
}

resource "opennebula_virtual_network" "network" {
  name   = "test-net1"
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

resource "opennebula_virtual_router_nic" "nic" {
  virtual_router_id = opennebula_virtual_router.test.id
  network_id        = opennebula_virtual_network.network2.id
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

To import an existing virtual router #42 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_virtual_router_nic" "importvr" {
    name = "importedvr"
}
```

And then run:

```
terraform import opennebula_virtual_router_nic.importvr 42
```

Verify that Terraform does not perform any change:

```
terraform plan
```
