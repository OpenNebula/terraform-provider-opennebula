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
resource "opennebula_virtual_machine" "example" {
  count = 2

  name        = "virtual-machine-${count.index}"
  description = "VM"
  cpu         = 1
  vcpu        = 1
  memory      = 1024
  group       = "terraform"
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
    boot = "disk0"
  }

  disk {
    image_id = opennebula_image.example.id
    size     = 10000
    target   = "vda"
    driver   = "qcow2"
  }

  on_disk_change = "RECREATE"

  nic {
    model           = "virtio"
    network_id      = var.vnetid
    security_groups = [opennebula_security_group.example.id]

    # NOTE: To make this work properly ensure /etc/one/oned.conf contains: INHERIT_VNET_ATTR="DNS"
    dns = "1.1.1.1"
  }

  vmgroup {
    vmgroup_id = 42
    role       = "vmgroup-role"
  }

  sched_requirements = "FREE_CPU > 60"

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

* `name` - (Required) The name of the virtual machine.
* `description`: (Optional) The description of the template.
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
* `keep_nic_order` - (Optional) Indicates if the provider should keep NIC list ordering at update.
* `vmgroup` - (Optional) See [VM group parameters](#vm-group-parameters) below for details. Changing this argument triggers a new resource.
* `group` - (Optional) Name of the group which owns the virtual machine. Defaults to the caller primary group.
* `raw` - (Optional) Allow to pass hypervisor level tuning content. See [Raw parameters](#raw-parameters) below for details.
* `sched_requirements` - (Optional) Scheduling requirements to deploy the resource following specific rule.
* `sched_ds_requirements` - (Optional) Storage placement requirements to deploy the resource following specific rule.
* `tags` - (Optional) Map of tags (`key=value`) assigned to the resource. Override matching tags present in the `default_tags` atribute when configured in the `provider` block. See [tags usage related documentation](https://registry.terraform.io/providers/OpenNebula/opennebula/latest/docs#using-tags) for more information.
* `timeout` - (Deprecated) Timeout (in Minutes) for VM availability. Defaults to 3 minutes.
* `lock` - (Optional) Lock the VM with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.
* `on_disk_change` - (Optional) Select the behavior for changing disk images. Supported values: `RECREATE` or `SWAP` (default). `RECREATE` forces recreation of the vm and `SWAP` adopts the standard behavior of hot-swapping the disks. NOTE: This property does not affect the behavior of adding new disks.
* `hard_shutdown` - (Optional) If the VM doesn't have ACPI support, it immediately poweroff/terminate/reboot/undeploy the VM. Defaults to false.
* `template_section` - (Optional) Allow to add a custom vector. See [Template section parameters](#template-section-parameters)

### Graphics parameters

`graphics` supports the following arguments:

* `type` - (Required) Generally set to VNC.
* `listen` - (Required) Binding address.
* `port` - (Optional) Binding Port.
* `keymap` - (Optional) Keyboard mapping.
* `passwd` - (Optional) VNC's password, conflicts with random_passwd.
* `random_passwd` - (Optional) Randomized VNC's password, conflicts with passwd.

### OS parameters

`os` supports the following arguments:

* `arch` - (Required) Hardware architecture of the Virtual machine. Must fit the host architecture.
* `boot` - (Optional) `OS` disk to use to boot on.
* `machine` - (Optional) libvirt machine type.
* `kernel` - (Optional) Path to the `OS` kernel to boot the image in the host.
* `kernel_ds` - (Optional) Image to be used as kernel. (see !!)
* `initrd` - (Optional) Path to the initrd image in the host.
* `initrd_ds` - (Optional) Image to be used as ramdisk. (see !!)
* `root` - (Optional) Device to be mounted as root.
* `kernel_cmd` - (Optional) Arguments for the booting kernel.
* `bootloader` - (Optional) Path to the bootloader executable.
* `sd_disk_bus` - (Optional) Bus for disks with sd prefix, either `scsi` or `sata`, if attribute is missing, libvirt chooses itself.
* `uuid` - (Optional) Unique ID of the VM.
* `firmware` - (Optional) Firmware type or firmware path. Possible values: `BIOS` or path for KVM, `BIOS` or `UEFI` for vCenter.
* `firmware_secure` - (Optional) Enable Secure Boot. (Can be `true` or `false`).
* (!!) Use one of `kernel_ds` or `kernel` (and `initrd` or `initrd_ds`).

### Disk parameters

`disk` supports the following arguments

* `image_id` - (Optional) ID of the image to attach to the virtual machine. Defaults to -1 if not set: this skip Image attchment to the VM. Conflicts with `volatile_type` and `volatile_format`.
* `size` - (Optional) Size (in MB) of the image. If set, it will resize the image disk to the targeted size. The size must be greater than the current one.
* `target` - (Optional) Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `driver` - (Optional) OpenNebula image driver.
* `volatile_type` - (Optional) Type of the disk: `swap` or `fs`. Type `swap` is not supported in vcenter. Conflicts with `image_id`.
* `volatile_format` - (Optional) Format of the Image: `raw` or `qcow2`. Conflicts with `image_id`.

Minimum 1 item. Maximum 8 items.

A disk update will be triggered in adding or removing a `disk` section, or by a modification of any of these parameters: `image_id`, `target`, `driver`

### NIC parameters

`nic` supports the following arguments

* `network_id` - (Optional) ID of the virtual network to attach to the virtual machine.
* `ip` - (Optional) IP of the virtual machine on this network.
* `mac` - (Optional) MAC of the virtual machine on this network.
* `model` - (Optional) Nic model driver. Example: `virtio`.
* `virtio_queues` - (Optional) Virtio multi-queue size. Only if `model` is `virtio`.
* `physical_device` - (Optional) Physical device hosting the virtual network.
* `security_groups` - (Optional) List of security group IDs to use on the virtual network.
* `method` - (Optional) Method of obtaining IP addresses (empty or `static`, `dhcp`, `skip`).
* `gateway` - (Optional) Default gateway set for the NIC.
* `dns` - (Optional) DNS server set for the NIC. **Please make sure `INHERIT_VNET_ATTR="DNS"` is added to `/etc/one/oned.conf`.**
* `network_mode_auto` - (Optional) A boolean letting the scheduler pick the Virtual Networks the VM NICs will be attached to. Can only be used at VM creation.
* `sched_requirements` - (Optional) A boolean expression to select virtual networks (evaluates to true) to attach the NIC,  when `network_mode_auto` is true. Can only be used at VM creation.
* `sched_rank` - (Optional) Arithmetic expression to sort the suitable Virtual Networks for this NIC, when `network_mode_auto` is true. Can only be used at VM creation.

Minimum 1 item. Maximum 8 items.

A NIC update will be triggered in adding or removing a `nic` section, or by a modification of any of these parameters: `network_id`, `ip`, `mac`, `security_groups`, `physical_device`

### VM group parameters

`vmgroup` supports the following arguments:

* `vmgroup_id` - (Required) ID of the VM group to use.
* `role` - (Required) role of the VM group to use.

### Raw parameters

`raw` supports the following arguments:

* `type` - (Required) - Hypervisor. Supported values: `kvm`, `lxd`, `vmware`.
* `data` - (Required) - Raw data to pass to the hypervisor.

### Template section parameters

`template_section` supports the following arguments:

* `name` - (Optional) The vector name.
* `elements` - (Optional) Collection of custom tags.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the virtual machine.
* `uid` - User ID whom owns the virtual machine.
* `gid` - Group ID which owns the virtual machine.
* `uname` - User Name whom owns the virtual machine.
* `gname` - Group Name which owns the virtual machine.
* `state` - State of the virtual machine.
* `lcmstate` - LCM State of the virtual machine.
* `template_disk` - when `template_id` is used and the template define some disks, this contains the template disks description.
* `template_nic` - when `template_id` is used and the template define some NICs, this contains the template NICs description.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.
* `template_tags` - When `template_id` was set this keeps the template tags.
* `template_section_names` - When `template_id` was set this keeps the template section names only.

### Template NIC

* `network_id` - ID of the image attached to the virtual machine.
* `nic_id` - nic attachment identifier
* `network` - network name
* `computed_ip` - IP of the virtual machine on this network.
* `computed_mac` - MAC of the virtual machine on this network.
* `computed_model` - Nic model driver.
* `computed_virtio_queues` - Virtio multi-queue size.
* `computed_physical_device` - Physical device hosting the virtual network.
* `computed_security_groups` - List of security group IDs to use on the virtual.
* `computed_method` - Method of obtaining IP addresses (empty or `static`, `dhcp`, `skip`).
* `computed_gateway` - Default gateway set for the NIC.
* `computed_dns` - DNS server set for the NIC.

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
* `computed_virtio_queues` - Virtio multi-queue size.
* `computed_physical_device` - Physical device hosting the virtual network.
* `computed_security_groups` - List of security group IDs to use on the virtual.
* `computed_method` - Method of obtaining IP addresses (empty or `static`, `dhcp`, `skip`).
* `computed_gateway` - Default gateway set for the NIC.
* `computed_dns` - DNS server set for the NIC.

### Disk

* `disk_id` - disk attachment identifier
* `computed_size` - Size (in MB) of the image attached to the virtual machine. Not possible to change a cloned image size.
* `computed_target` - Target name device on the virtual machine. Depends of the image `dev_prefix`.
* `computed_driver` - OpenNebula image driver.
* `computed_volatile_format` - Format of the Image: `raw` or `qcow2`.

## Instantiate from a template

When the attribute `template_id` is set, here is the behavior:

For all parameters excepted context: parameters present in VM overrides parameters defined in template.
For context: it merges them.

For disks and NICs defined in the template, if they are not overriden, are described in `template_disk` and `template_nic` attributes of the instantiated VM and are not modifiable anymore.

## Import

`opennebula_virtual_machine` can be imported using its ID:

```shell
terraform import opennebula_virtual_machine.example 123
```
