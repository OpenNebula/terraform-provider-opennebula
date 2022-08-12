---
layout: "opennebula"
page_title: "OpenNebula: opennebula_disk"
sidebar_current: "docs-opennebula-resource-disk"
description: |-
  Provides an OpenNebula virtual machine disk resource.
---

# opennebula_disk

Provides an OpenNebula virtual machine disk resource.

This resource allows you to manage virtual machines disks. When applied,
a new disk is attached to the virtual machine. When destroyed, the disk is detached from the virtual machine.

## Example Usage

```hcl

resource "opennebula_disk" "example1" {
    image_id = opennebula_image.example.id
    target   = "vdb"
}

resource "opennebula_disk" "example2" {
    volatile_type   = "swap"
    volatile_format = "raw"
    size            = 16
    target          = "vdc"
}

resource "opennebula_image" "example" {
	name             = "example"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
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

  disk  {
    image_id = opennebula_image.example.id
    target   = "vda"
  }

  lifecycle {
      ignore_changes = [
          disk,
      ]
  }

  hard_shutdown = true
}
```

## Argument Reference

`disk` supports the following arguments

* `vm_id` - (Required) ID of the virtual machine.
* `image_id` - (Optional) ID of the image to attach to the virtual machine. Defaults to -1 if not set: this skip Image attchment to the VM. Conflicts with `volatile_type` and `volatile_format`.
* `size` - (Optional) Size (in MB) of the image. If set, it will resize the image disk to the targeted size. The size must be greater than the current one.
* `target` - (Optional) Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.
* `volatile_type` - (Optional) Type of the disk: `swap` or `fs`. Type `swap` is not supported in vcenter. Conflicts with `image_id`.
* `volatile_format` - (Optional) Format of the Image: `raw` or `qcow2`. Conflicts with `image_id`.

## Import


`opennebula_disk` can be imported using a composed ID:

```sh
terraform import opennebula_disk.example vm_id:disk_id
```