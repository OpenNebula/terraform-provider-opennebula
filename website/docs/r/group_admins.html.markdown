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
data "template_file" "tpl" {
  template = file("group_template.txt")
}

resource "opennebula_group" "group" {
  name                  = "test_group"
  template              = data.template_file.tpl.rendered
}

resource "opennebula_user" "user" {
  name          = "test_user"
  password      = "p@ssw0rd"
  auth_driver   = "core"
  primary_group = opennebula_group.group.id
}

resource "opennebula_group_admins" "admins" {
  group_id = opennebula_group.group.id
  users_ids = [
    opennebula_user.user.id
  ]
}
```

with `group_template.txt` file with Sunstone information:

```php
SUNSTONE = [
  DEFAULT_VIEW = "cloud",
  group_admins_ADMIN_DEFAULT_VIEW = "group_adminsadmin",
  group_admins_ADMIN_VIEWS = "cloud,group_adminsadmin",
  VIEWS = "cloud"
]
```

## Argument Reference

The following arguments are supported:

* `groups_id` - (Required) The id of the related group.
* `users_ids` - (Required) List of users ids

## Import

To import an existing group_admins #134 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_group_admins" "import_group_admins" {
    name = "importedgroup_admins"
}
```

And then run:

```
terraform import opennebula_group_admins.import_group_admins 134
```

Verify that Terraform does not perform any change:

```
terraform plan
```
