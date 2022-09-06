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
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_virtual_router` can be imported using its ID:

```sh
terraform import opennebula_virtual_router.example 123
```
