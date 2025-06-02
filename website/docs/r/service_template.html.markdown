---
layout: "opennebula"
page_title: "OpenNebula: opennebula_service_template"
sidebar_current: "docs-opennebula-resource-service-template"
description: |-
  Provides an OpenNebula service template resource.
---

# opennebula_service_template

Provides an OpenNebula service template resource.

This resource allows you to manage service templates on your OpenNebula clusters. When applied,
a new service template will be created. When destroyed, that service template will be removed.

## Example Usage

```hcl
resource "opennebula_service_template" "example" {
  name        = "servicetemplate"
  template    = templatefile("${path.module}/template.json", {})
  permissions = 642
  uname       = "oneadmin"
  gname       = "oneadmin"
}
```

`template.json` file contains a `json` document with the definition of the service template:

```json
{
  "TEMPLATE": {
    "BODY": {
      "name": "tm-stemplate",
      "deployment": "straight",
      "roles": [
        {
          "name": "master",
          "cardinality": 3,
          "type": "vm",
          "template_id": 0,
          "min_vms": 2
        }
      ]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the service template.
* `permissions` - (Optional) Permissions applied on service template. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `template` - (Optional) Service template definition in JSON format.
* `uid` - (Optional) Set the id of the user owner of the newly created service template. The corresponding `uname` will be computed.
* `uname` - (Optional) Set the name of the user owner of the newly created service template. The corresponding `uid` will be computed.
* `gid` - (Optional) Set the id of the group owner of the newly created service template. The corresponding `gname` will be computed.
* `gname` - (Optional) Set the name of the group owner of the newly created service template. The corresponding `gid` will be computed.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the service.
* `uid` - User ID whom owns the service.
* `gid` - Group ID which owns the service.
* `uname` - User Name whom owns the service.
* `gname` - Group Name which owns the service.

## Import

`opennebula_security_group` can be imported using its ID:

```shell
terraform import opennebula_service_template.example 123
```
