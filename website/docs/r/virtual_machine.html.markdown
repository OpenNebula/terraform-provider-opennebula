---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_machine"
sidebar_current: "docs-opennebula-resource-virtual-machine"
description: |-
  Provides an OpenNebula virtual machine resource.
---

# opennebula_virtual_machine

Provides an OpenNebula virtual machine resource.

This resource allows you to manage virtual machines on your OpenNebula clusters. When applied,
a new virtual machine will be created. When destroyed, that virtual machine will be removed.

## Example Usage

```hcl
data "template_file" "cloudinit" {
  template = "${file("cloud-init.yaml")}"
}

resource "opennebula_virtual_machine" "demo" {
  count = 2
  name = "tfdemovm"
  cpu = 1
  vcpu = 1
  memory = 1024
  group = "terraform"
  permissions = "660"

  context {
    NETWORK = "YES"
    HOSTNAME = "$NAME"
    USER_DATA = "${data.template_file.cloudinit.rendered}"
  }

  graphics {
    type = "VNC"
    listen = "0.0.0.0"
    keymap = "fr"
  }

  os {
    arch = "x86_64"
    boot = "disk0"
  }

  disk {
    image_id = "${opennebula_image.osimage.id}"
    size = 10000
    target = "vda"
    driver = "qcow2"
  }

  nic {
    model = "virtio-pci-net"
    network_id = "${var.vnetid}"
    security_groups = ["${opennebula_security_group.mysecgroup.id}"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual machine.
* `permissions` - (Optional) Permissions applied on virtual machine. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin.
* `template_id` - (Optional) If set, VM are instantiated from the template ID.
* `pending` - (Optional) Pending state during VM creation. Defaults to `false`.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM. **Mandatory if `template_****id` is not set**.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB. **Mandatory if `template_****id` is not set**.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include at a minimum: NETWORK = "YES" and SET_HOSTNAME = "$NAME.
* `graphics` - (Optional) See [Graphics parameters](#graphics-vm) below for details.
* `os` - (Optional) See [OS parameters](#os-vm) below for details.
* `disk` - (Optional) Can be specified multiple times to attach several disks. See [Disks parameters](#disks-vm) below for details.
* `nic` - (Optional) Can be specified multiple times to attach several NICs. See [Nic parameters](#nic-vm) below for details.
* `group` - (Optional) Name of the group which owns the virtual machine. Defaults to the caller primary group.

### Graphics parameters

`graphics` supports the following arguments:

* `type` - (Required) Generally set to VNC.
* `listen` - (Required) Binding address.
* `port` - (Optional) Binding Port.
* `keymap` - (Optional) Keyboard mapping.

### OS parameters

`os` supports the following arguments:

* `arch` - (Required) Hardware architecture of the Virtual machine. Must fit the host architecture.
* `boot` - (Optional) `OS` disk to use to boot on.

### Disk parameters

`disk` supports the following arguments

* `image_id` - (Required) ID of the image to attach to the virtual machine.
* `size` - (Optional) Size (in MB) of the image attached to the virtual machine. Not possible to change a cloned image size.
* `target` - (Optional) Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.

Minimum 1 item. Maximum 8 items.

### NIC parameters

`nic` supports the following arguments

* `network_id` - (Required) ID of the virtual network to attach to the virtual machine.
* `ip` - (Optional) IP of the virtual machine on this network.
* `mac` - (Optional) MAC of the virtual machine on this network.
* `model` - (Optional) Nic model driver. Example: `virtio`.
* `physical_device` - (Optional) Physical device hosting the virtual network.
* `security_groups` - (Optional) List of security group IDs to use on the virtual network.

Minimum 1 item. Maximum 8 items.

## Attribute Reference

The following attribute are exported:
* `id` - ID of the virtual machine.
* `instance` - (Deprecated) Name of the virtual machine instance created on the cluster.
* `uid` - User ID whom owns the virtual machine.
* `gid` - Group ID which owns the virtual machine.
* `uname` - User Name whom owns the virtual machine.
* `gname` - Group Name which owns the virtual machine.
* `state` - State of the virtual machine.
* `lcmstate` - LCM State of the virtual machine.

## Import

To import an existing virtual machine #42 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_virtual_machine" "importvm" {
    name = "importedvm"
}

And then run:

```
terraform import opennebula_virtual_machine.importvm 42
```

Verify that Terraform does not perform any change:

```
terraform plan
```

