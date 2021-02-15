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
a new virtual machine is created. When destroyed, this virtual machine is removed.

## Example Usage

```hcl

resource "opennebula_virtual_machine" "demo" {
  count       = 2
  name        = "tfdemovm"
  cpu         = 1
  vcpu        = 1
  memory      = 1024
  group       = "terraform"
  permissions = "660"

  context {
    NETWORK      = "YES"
    HOSTNAME     = "$NAME"
    START_SCRIPT ="yum upgrade"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "fr"
  }

  os {
    arch = "x86_64"
    boot = "disk0"
  }

  disk {
    image_id = opennebula_image.osimage.id
    size     = 10000
    target   = "vda"
    driver   = "qcow2"
  }

  nic {
    model           = "virtio"
    network_id      = var.vnetid
    security_groups = [opennebula_security_group.mysecgroup.id]
  }

  vmgroup {
    vmgroup_id = 42
    role       = "vmgroup-role"
  }

  tags = {
    environment = "dev"
  }

  timeout = 5
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual machine.
* `permissions` - (Optional) Permissions applied on virtual machine. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `template_id` - (Optional) If set, VM are instantiated from the template ID. See [Instantiate from a template](#instantiate-from-a-template) for details. Changing this argument triggers a new resource.
* `pending` - (Optional) Pending state during VM creation. Defaults to `false`.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM. **Mandatory if** `template_id` **is not set**.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB. **Mandatory if** `template_id` **is not set**.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include: `NETWORK = "YES"` and `SET_HOSTNAME = "$NAME"`. If a `template_id` is set, see [Instantiate from a template](#instantiate-from-a-template) for details.
* `graphics` - (Optional) See [Graphics parameters](#graphics-parameters) below for details.
* `os` - (Optional) See [OS parameters](#os-parameters) below for details.
* `disk` - (Optional) Can be specified multiple times to attach several disks. See [Disk parameters](#disk-parameters) below for details.
* `nic` - (Optional) Can be specified multiple times to attach several NICs. See [Nic parameters](#nic-parameters) below for details.
* `vmgroup` - (Optional) See [VM group parameters](#vm-group-parameters) below for details. Changing this argument triggers a new resource.
* `group` - (Optional) Name of the group which owns the virtual machine. Defaults to the caller primary group.
* `tags` - (Optional) Virtual Machine tags (Key = Value).
* `timeout` - (Optional) Timeout (in Minutes) for VM availability. Defaults to 3 minutes.

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

* `image_id` - (Optional) ID of the image to attach to the virtual machine. Defaults to -1 if not set: this skip Image attchment to the VM.
* `size` - (Optional) Size (in MB) of the image. If set, it will resize the image disk to the targeted size. The size must be greater than the current one.
* `target` - (Optional) Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.

Minimum 1 item. Maximum 8 items.

A disk update will be triggered in adding or removing a `disk` section, or by a modification of any of these parameters: `image_id`, `target`, `driver`

### NIC parameters

`nic` supports the following arguments

* `network_id` - (Required) ID of the virtual network to attach to the virtual machine.
* `ip` - (Optional) IP of the virtual machine on this network.
* `mac` - (Optional) MAC of the virtual machine on this network.
* `model` - (Optional) Nic model driver. Example: `virtio`.
* `physical_device` - (Optional) Physical device hosting the virtual network.
* `security_groups` - (Optional) List of security group IDs to use on the virtual network.

Minimum 1 item. Maximum 8 items.

A NIC update will be triggered in adding or removing a `nic` section, or by a modification of any of these parameters: `network_id`, `ip`, `mac`, `security_groups`, `physical_device`

### VM group parameters

`vmgroup` supports the following arguments:

* `vmgroup_id` - (Required) ID of the VM group to use.
* `role` - (Required) role of the VM group to use.

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
* `template_disk` - when `template_id` is used and the template define some disks, this contains the template disks description.
* `template_nic` - when `template_id` is used and the template define some NICs, this contains the template NICs description.


### Template NIC

* `network_id` - ID of the image attached to the virtual machine.
* `nic_id` - nic attachment identifier
* `network` - network name
* `computed_ip` - IP of the virtual machine on this network.
* `computed_mac` - MAC of the virtual machine on this network.
* `computed_model` - Nic model driver.
* `computed_physical_device` - Physical device hosting the virtual network.
* `computed_security_groups` - List of security group IDs to use on the virtual.


### Template disk

* `image_id` - ID of the image attached to the virtual machine.
* `disk_id` - disk attachment identifier
* `computed_size` - Size (in MB) of the image attached to the virtual machine. Not possible to change a cloned image size.
* `computed_target` - Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `computed_driver` - OpenNebula image driver.

### NIC

* `nic_id` - nic attachment identifier
* `network` - network name
* `computed_ip` - IP of the virtual machine on this network.
* `computed_mac` - MAC of the virtual machine on this network.
* `computed_model` - Nic model driver.
* `computed_physical_device` - Physical device hosting the virtual network.
* `computed_security_groups` - List of security group IDs to use on the virtual.

### Disk

* `disk_id` - disk attachment identifier
* `computed_size` - Size (in MB) of the image attached to the virtual machine. Not possible to change a cloned image size.
* `computed_target` - Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `computed_driver` - OpenNebula image driver.

## Instantiate from a template

When the attribute `template_id` is set, here is the behavior:

For all parameters excepted context: parameters present in VM overrides parameters defined in template.
For context: it merges them.

For disks and NICs defined in the template, if they are not overriden, are described in `template_disk` and `template_nic` attributes of the instantiated VM and are not modifiable anymore.

## Import

To import an existing virtual machine #42 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_virtual_machine" "importvm" {
    name = "importedvm"
}
```

And then run:

```
terraform import opennebula_virtual_machine.importvm 42
```

Verify that Terraform does not perform any change:

```
terraform plan
```

