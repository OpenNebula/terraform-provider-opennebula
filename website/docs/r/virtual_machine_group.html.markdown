---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_machine_group"
sidebar_current: "docs-opennebula-resource-virtual-machine-group"
description: |-
  Provides an OpenNebula virtual machine group resource.
---

# opennebula_virtual_machine_group

Provides an OpenNebula virtual machine group resource.

This resource allows you to manage virtual machine groups on your OpenNebula clusters. When applied,
a new virtual machine group will be created. When destroyed, that virtual machine group will be removed.

## Example Usage

```hcl
resource "opennebula_virtual_machine_group" "test" {
  name        = "test-vmgroup"
  group       = "oneadmin"
  permissions = "642"
  role {
    name = "anti-aff"
    host_anti_affined = [ 0 ]
    policy = "ANTI_AFFINED"
  }
  tags = {
    environment = "dev"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual machine group.
* `permissions` - (Optional) Permissions applied on virtual machine group. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin.
* `role` - (Required) List of roles. See [Role parameters](#role-parameters) below for details.
* `group` - (Optional) Name of the group which owns the virtual machine group. Defaults to the caller primary group.
* `tags` - (Optional) Virtual Machine group tags.

### Role parameters

`role` supports the following arguments:

* `name` - (Required) Name of the role.
* `host_affined` - (Optional) List of Hosts affined to Virtual Machines using this role.
* `host_anti_affined` - (Optional) List of Hosts not-affined to Virtual Machines using this role.
* `policy` - (Optional) Policy to apply between Virtual Machines using this role. Allowed Values: `NONE`, `AFFINED`, `ANTI_AFFINED`.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the virtual machine.
* `uid` - User ID whom owns the virtual machine.
* `gid` - Group ID which owns the virtual machine.
* `uname` - User Name whom owns the virtual machine.
* `gname` - Group Name which owns the virtual machine.
* `role` - See [Role Attribute Reference](#role-attribute-reference) below for details

## Role Attribute Reference

The Following attributes are exported under `role`:

* `id` - ID of the role.
* `vms` - List of Virtual Machine IDs using this role.

## Import

To import an existing virtual machine group #42 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_virtual_machine_group" "importvmg" {
    name = "importedvmg"
}
```

And then run:

```
terraform import opennebula_virtual_machine_group.importvmg 42
```

Verify that Terraform does not perform any change:

```
terraform plan
```

