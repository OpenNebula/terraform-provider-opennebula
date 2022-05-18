---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_router"
sidebar_current: "docs-opennebula-resource-virtual-router"
description: |-
  Provides an OpenNebula virtual router resource.
---

# opennebula_virtual_router

Provides an OpenNebula virtual router resource.

## Example Usage

```hcl
resource "opennebula_virtual_router" "vrouter" {
  name        = "testacc-vr"
  permissions = "642"
  group       = "oneadmin"
  description = "This is an example of virtual router"

  instance_template_id = opennebula_virtual_router_instance_template.test.id

  lock        = "USE"
  tags = {
    environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual router.
* `instance_template_id` - (Required) The ID of the template of the virtual router instances.
* `permissions` - (Optional) Permissions applied on virtual router. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `group` - (Optional) Name of the group which owns the virtual router. Defaults to the caller primary group.
* `description` - (Optional) Description of the virtual router.
* `lock` - (Optional) Lock the VM with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the virtual router.
* `uid` - User ID whom owns the virtual router.
* `gid` - Group ID which owns the virtual router.
* `uname` - User Name whom owns the virtual router.
* `gname` - Group Name which owns the virtual router.

## Import

To import an existing virtual router #42 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_virtual_router" "importvr" {
    name = "importedvr"
}
```

And then run:

```sh
terraform import opennebula_virtual_router.importvr 42
```

Verify that Terraform does not perform any change:

```sh
terraform plan
```
