---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_router_instance_template"
sidebar_current: "docs-opennebula-resource-virtual-router"
description: |-
  Provides an OpenNebula virtual router resource.
---

# opennebula_virtual_router_instance template

Provides an OpenNebula virtual router instance template resource.

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
```

## Argument Reference

* `name` - (Required) The name of the virtual machine template.
* `description`: (Optional) The description of the template.
* `permissions` - (Optional) Permissions applied on template. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `group` - (Optional) Name of the group which owns the template. Defaults to the caller primary group.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `features` - (Optional) See [Features parameters](#features-parameters) below for details.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include: `NETWORK = "YES"` and `SET_HOSTNAME = "$NAME"`.
* `graphics` - (Optional) See [Graphics parameters](#graphics-parameters) below for details.
* `os` - (Optional) See [OS parameters](#os-parameters) below for details.
* `disk` - (Optional) Can be specified multiple times to attach several disks. See [Disks parameters](#disks-parameters) below for details.
* `raw` - (Optional) Allow to pass hypervisor level tuning content. See [Raw parameters](#raw-parameters) below for details.
* `vmgroup` - (Optional) See [VM group parameters](#vm-group-parameters) below for details. Changing this argument triggers a new resource.
* `user_inputs` - (Optional) Ask the user instantiating the template to define the values described.
* `sched_requirements` - (Optional) Scheduling requirements to deploy the resource following specific rule
* `sched_ds_requirements` - (Optional) Storage placement requirements to deploy the resource following specific rule.
* `tags` - (Optional) Template tags (Key = Value).
* `lock` - (Optional) Lock the template with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.

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

### Features parameters

`features` supports the following arguments

* `pea` - (Optional) Physical address extension mode allows 32-bit guests to address more than 4 GB of memory. (Can be `YES` or `NO`)
* `acpi` - (Optional) Useful for power management, for example, with KVM guests it is required for graceful shutdown to work. (Can be `YES` or `NO`)
* `apic` - (Optional) Enables the advanced programmable IRQ management. Useful for SMP machines. (Can be `YES` or `NO`)
* `localtime` - (Optional) The guest clock will be synchronized to the host’s configured timezone when booted. Useful for Windows VMs. (Can be `YES` or `NO`)
* `hyperv` - (Optional) Add hyperv extensions to the VM. The options can be configured in the driver configuration, HYPERV_OPTIONS.
* `guest_agent` - (Optional) Enables the QEMU Guest Agent communication. This only creates the socket inside the VM, the Guest Agent itself must be installed and started in the VM. (Can be `YES` or `NO`)
* `virtio_scsi_queues` - (Optional) Numer of vCPU queues for the virtio-scsi controller.
* `iothreads` - (Optional) umber of iothreads for virtio disks. By default threads will be assign to disk by round robin algorithm. Disk thread id can be forced by disk IOTHREAD attribute.

### Disk parameters

`disk` supports the following arguments

* `image_id` - (Optional) ID of the image to attach to the virtual machine. Conflicts with `volatile_type` and `volatile_format`.
* `size` - (Optional) Size (in MB) of the image attached to the virtual machine. Not possible to change a cloned image size.
* `target` - (Optional) Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.
* `volatile_type` - (Optional) Type of the volatile disk: `swap` or `fs`. Type `swap` is not supported in vcenter. Conflicts with `image_id`.
* `volatile_format` - (Optional) Format of the volatile disk: `raw` or `qcow2`. Conflicts with `image_id`.

Minimum 1 item. Maximum 8 items.

### Raw parameters

`raw` supports the following arguments:

* `type` - (Required) - Hypervisor. Supported values: `kvm`, `lxd`, `vmware`.
* `data` - (Required) - Raw data to pass to the hypervisor.

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
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_virtual_router_instance_template` can be imported using its ID:

```shell
terraform import opennebula_virtual_router_instance_template.example 123
```
