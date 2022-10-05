---
layout: "opennebula"
page_title: "OpenNebula: opennebula_datastore"
sidebar_current: "docs-opennebula-datasource-datastore"
description: |-
  Get the datastore information for a given name.
---

# opennebula_datastore

Use this data source to retrieve the datastore information from it's name or tags.

## Example Usage

```hcl
data "opennebula_datastore" "example" {
  name = "My_Datastore"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula datastore to retrieve information for.
* `tags` - (Optional) Tags associated to the datastore.

## Attribute Reference

* `id` - ID of the datastore.
* `name` - The OpenNebula datastore name.
* `tags` - Tags of the datastore (Key = Value).
