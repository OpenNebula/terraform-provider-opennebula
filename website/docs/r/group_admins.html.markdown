---
layout: "opennebula"
page_title: "OpenNebula: opennebula_group_admins"
sidebar_current: "docs-opennebula-resource-group_admins"
description: |-
  Provides an OpenNebula group_admins resource.
---

# opennebula_group_admins

Provides an OpenNebula group administrators resource.

This resource allows you to manage group administrators on OpenNebula. When applied,
adminstrator are added or removed from the group. When destroyed, all adminstrators are removed from the group.

## Example Usage

```hcl
data "template_file" "example" {
  template = file("group_template.txt")
}

resource "opennebula_group" "example" {
  name     = "group"
  template = data.template_file.example.rendered
}

resource "opennebula_user" "example" {
  name          = "user"
  password      = "p@ssw0rd"
  auth_driver   = "core"
  primary_group = opennebula_group.example.id
}

resource "opennebula_group_admins" "example" {
  group_id = opennebula_group.example.id
  users_ids = [
    opennebula_user.example.id
  ]
}
```

with `group_template.txt` file with Sunstone information:

```raw
SUNSTONE = [
  DEFAULT_VIEW = "cloud",
  group_admins_ADMIN_DEFAULT_VIEW = "group_adminsadmin",
  group_admins_ADMIN_VIEWS = "cloud,group_adminsadmin",
  VIEWS = "cloud"
]
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required) The id of the related group.
* `users_ids` - (Required) List of users ids

## Import

`opennebula_group_admins` can be imported using the group ID:

```shell
terraform import opennebula_group_admins.example 123
```
