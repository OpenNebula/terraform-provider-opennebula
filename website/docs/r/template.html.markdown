---
layout: "opennebula"
page_title: "OpenNebula: opennebula_template"
sidebar_current: "docs-opennebula-resource-template"
description: |-
  Provides an OpenNebula template resource.
---

# opennebula_template

Provides an OpenNebula template resource.

This resource allows you to manage templates on your OpenNebula clusters. When applied,
a new template will be created. When destroyed, that template will be removed.

## Example Usage

```hcl
resource "opennebula_template" "mytemplate" {
  name = "mytemplate"
  cpu = 1
  vcpu = 1
  memory = 1024
  group = "terraform"
  permissions = "660"

  context {
    NETWORK = "YES"
    HOSTNAME = "$NAME"
    START_SCRIPT="yum upgrade"
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

  vmgroup {
    vmgroup_id = 42
    role = "vmgroup-role"
  }

  tags = {
    environment = "dev"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual machine template.
* `permissions` - (Optional) Permissions applied on template. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin.
* `group` - (Optional) Name of the group which owns the template. Defaults to the caller primary group.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM. **Mandatory if `template_****id` is not set**.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB. **Mandatory if `template_****id` is not set**.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include at a minimum: NETWORK = "YES" and SET_HOSTNAME = "$NAME.
* `graphics` - (Optional) See [Graphics parameters](#graphics-parameters) below for details.
* `os` - (Optional) See [OS parameters](#os-parameters) below for details.
* `disk` - (Optional) Can be specified multiple times to attach several disks. See [Disks parameters](#disks-parameters) below for details.
* `nic` - (Optional) Can be specified multiple times to attach several NICs. See [Nic parameters](#nic-parameters) below for details.
* `vmgroup` - (Optional) See [VM group parameters](#os-parameters) below for details. Changing this argument triggers a new resource.
* `tags` - (Optional) Template tags.
* `template` - (Deprecated) Text describing the OpenNebula template object, in Opennebula's XML string format.

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

### VM group parameters

`vmgroup` supports the following arguments:

* `vmgroup_id` - (Required) ID of the VM group to use.
* `role` - (Required) role of the VM group to use.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the template.
* `uid` - User ID whom owns the template.
* `gid` - Group ID which owns the template.
* `uname` - User Name whom owns the template.
* `gname` - Group Name which owns the template.
* `reg_time` - Registration time of the template.

## Import

To import an existing virtual machine template #54 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_template" "importtpl" {
    name = "importedtpl"
}
```

And then run:

```
terraform import opennebula_template.importtppl 54
```

Verify that Terraform does not perform any change:

```
terraform plan
```
