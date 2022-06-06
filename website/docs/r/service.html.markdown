---
layout: "opennebula"
page_title: "OpenNebula: opennebula_service"
sidebar_current: "docs-opennebula-resource-service"
description: |-
  Provides an OpenNebula service resource.
---

# opennebula_service

Provides an OpenNebula service resource.

This resource allows you to manage services on your OpenNebula clusters. When applied,
a new service will be created. When destroyed, that service will be removed.

## Example Usage

```hcl
resource "opennebula_service" "example" {
  name           = "service"
  template_id    = 11
  extra_template = templatefile("${path.module}/extra_template.json", {})
}
```

`extra_template.json` file contains a `json` document with extra information used during service instantiate (e.g networks, custom attriubtes...):

```json
{
  "networks_values": [
    {
      "vnet": {
        "reserve_from": "2",
        "extra": "SIZE=3"
      }
    }
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the service.
* `permissions` - (Optional) Permissions applied on service. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `template_id` - (Required) Service will be instantiated from the template ID.
* `extra_template` - (Optional) Service information to be merged with the template during instantiate.
* `uid` - (Optional) Set the id of the user owner of the newly created service. The corresponding `uname` will be computed.
* `uname` - (Optional) Set the name of the user owner of the newly created service. The corresponding `uid` will be computed.
* `gid` - (Optional) Set the id of the group owner of the newly created service. The corresponding `gname` will be computed.
* `gname` - (Optional) Set the name of the group owner of the newly created service. The corresponding `gid` will be computed.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the service.
* `uid` - User ID whom owns the service.
* `gid` - Group ID which owns the service.
* `uname` - User Name whom owns the service.
* `gname` - Group Name which owns the service.
* `state` - State of the service.
* `networks` - Map with the service name of each networks along with the id of the network.
* `roles` - Array with roles information containing: `cardinality`, `name`, `nodes` and `state`.

## Import

`opennebula_service` can be imported using its ID:

```shell
terraform import opennebula_service.example 123
```
