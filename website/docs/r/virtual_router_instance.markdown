---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_router_instance"
sidebar_current: "docs-opennebula-resource-virtual-router-instance"
description: |-
  Provides an OpenNebula virtual router instance resource.
---

# opennebula_virtual_router_instance

Provides an OpenNebula virtual router instance resource.

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
    environment = "example"
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
```

## Argument Reference

The following arguments are supported:

* `virtual_router_id` - (Required) The ID of the parent virtual router resource.
* `name` - (Required) The name of the virtual router instance.
* `description`: (Optional) The description of the template.
* `permissions` - (Optional) Permissions applied on virtual router instance. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `pending` - (Optional) Pending state during VM creation. Defaults to `false`.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include: `NETWORK = "YES"` and `SET_HOSTNAME = "$NAME"`.
* `graphics` - (Optional) See [Graphics parameters](#graphics-parameters) below for details.
* `os` - (Optional) See [OS parameters](#os-parameters) below for details.
* `disk` - (Optional) Can be specified multiple times to attach several static disks at VM creation. Disk described here won't be updated. See `opennebula_disk` resource for flexible disk management. See [Disk parameters](#disk-parameters) below for details.
* `vmgroup` - (Optional) See [VM group parameters](#vm-group-parameters) below for details. Changing this argument triggers a new resource.
* `group` - (Optional) Name of the group which owns the virtual router instance. Defaults to the caller primary group.
* `sched_requirements` - (Optional) Scheduling requirements to deploy the resource following specific rule.
* `sched_ds_requirements` - (Optional) Storage placement requirements to deploy the resource following specific rule.
* `tags` - (Optional) virtual router instance tags (Key = Value).
* `lock` - (Optional) Lock the VM with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.
* `on_disk_change` - (Deprecated) Select the behavior for changing disk images. Supported values: `RECREATE` or `SWAP` (default). `RECREATE` forces recreation of the vm and `SWAP` adopts the standard behavior of hot-swapping the disks. NOTE: This property does not affect the behavior of adding new disks.

### Graphics parameters

`graphics` supports the following arguments:

* `type` - (Required) Generally set to VNC.
* `listen` - (Required) Binding address.
* `port` - (Optional) Binding Port.
* `keymap` - (Optional) Keyboard mapping.

### OS parameters

`os` supports the following arguments:

* `arch` - (Required) Hardware architecture of the virtual router instance. Must fit the host architecture.
* `boot` - (Optional) `OS` disk to use to boot on.

### Disk parameters

`disk` supports the following arguments

* `image_id` - (Optional) ID of the image to attach to the virtual router instance. Defaults to -1 if not set: this skip Image attchment to the VM. Conflicts with `volatile_type` and `volatile_format`.
* `size` - (Optional) Size (in MB) of the image. If set, it will resize the image disk to the targeted size. The size must be greater than the current one.
* `target` - (Optional) Target name device on the virtual router instance. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.
* `volatile_type` - (Optional) Type of the disk: `swap` or `fs`. Type `swap` is not supported in vcenter. Conflicts with `image_id`.
* `volatile_format` - (Optional) Format of the Image: `raw` or `qcow2`. Conflicts with `image_id`.

Minimum 1 item. Maximum 8 items.

### VM group parameters

`vmgroup` supports the following arguments:

* `vmgroup_id` - (Required) ID of the VM group to use.
* `role` - (Required) role of the VM group to use.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the virtual router instance.
* `uid` - User ID whom owns the virtual router instance.
* `gid` - Group ID which owns the virtual router instance.
* `uname` - User Name whom owns the virtual router instance.
* `gname` - Group Name which owns the virtual router instance.
* `state` - State of the virtual router instance.
* `lcmstate` - LCM State of the virtual router instance.
* `template_disk` - this contains the template disks description.

### Template disk

* `image_id` - ID of the image attached to the virtual router instance.
* `disk_id` - disk attachment identifier
* `computed_size` - Size (in MB) of the image attached to the virtual router instance. Not possible to change a cloned image size.
* `computed_target` - Target name device on the virtual router instance. Depends of the image `dev_prefix`.
* `computed_driver` - OpenNebula image driver.

### Disk

* `disk_id` - disk attachment identifier
* `computed_size` - Size (in MB) of the image attached to the virtual router instance. Not possible to change a cloned image size.
* `computed_target` - Target name device on the virtual router instance. Depends of the image `dev_prefix`.
* `computed_driver` - OpenNebula image driver.
* `computed_volatile_format` - Format of the Image: `raw` or `qcow2`.

## Instantiate from a template

A virtual router instance is created from a template.
The template ID is defined in the virtual router resource.

For all virtual router instance parameters excepted context: parameters present in instance overrides parameters defined in template.
For context: it merges them.

For disks defined in the template, if they are not overriden, are described in `template_disk` attributes of the instantiated virtual router instance and are not modifiable anymore.

## Import

`opennebula_virtual_router_instance` can be imported using its ID:

```sh
terraform import opennebula_virtual_router_instance.example 123
```
