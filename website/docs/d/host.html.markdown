---
layout: "opennebula"
page_title: "OpenNebula: opennebula_host"
sidebar_current: "docs-opennebula-datasource-host"
description: |-
  Get the host information for a given name.
---

# opennebula_host

Use this data source to retrieve the host information for a given name.

## Example Usage

```hcl
data "opennebula_host" "example" {
  name = "My_Host"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula host to retrieve information for.
* `tags` - (Optional) Tags associated to the host.

## Attribute Reference

* `id` - ID of the host.
* `name` - Name of the host.
* `tags` - Tags of the host (Key = Value).
