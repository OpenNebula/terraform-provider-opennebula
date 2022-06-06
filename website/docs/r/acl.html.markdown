---
layout: "opennebula"
page_title: "OpenNebula: opennebula_acl"
sidebar_current: "docs-opennebula-resource-acl"
description: |-
  Provides an OpenNebula acl resource.
---

# opennebula_acl

Provides an OpenNebula ACL resource.

This resource allows you to manage ACLs on your OpenNebula clusters. When applied,
a new ACL is created. When destroyed, this ACL is removed. Note that ACLs currently cannot be changed, hence they are deleted and re-created upon change.

## Example Usage

```hcl
resource "opennebula_acl" "example" {
  user     = "@1"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights   = "USE+MANAGE+ADMIN"
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) User component of the new rule.
  * `#<id>` matches a single user id
  * `@<id>` matches a group id
  * `*` matches everything.
* `resource` - (Required) Resource component of the new rule. Any combination of valid resources, separated by a `+`.

  **Must contain a slash for resource subset.**
  Resource subset string uses the same syntax as the User-string, and additionally supports `%<id>` to limit by Cluster ID.

  The following objects are valid:
  * VM
  * HOST
  * NET
  * IMAGE
  * USER
  * TEMPLATE
  * GROUP
  * DATASTORE
  * CLUSTER
  * DOCUMENT
  * ZONE
  * SECGROUP
  * VDC
  * VROUTER
  * MARKETPLACE
  * MARKETPLACEAPP
  * VMGROUP
  * VNTEMPLATE
* `rights` - (Optional) Rights component of the new rule. Any combination of valid Rights, separated by a `+`.

  The following rights are valid:
  * USE
  * MANAGE
  * ADMIN
  * CREATE
* `zone` - (Optional) Zone component of the new rule.
  * `#<id>` matches a single zone id
  * `*` matches everything.

## Import

`opennebula_acl` can be imported using its ID:

```shell
terraform import opennebula_acl.example 123
```
