---
layout: "opennebula"
page_title: "OpenNebula: opennebula_marketplace"
sidebar_current: "docs-opennebula-datasource-marketplace"
description: |-
  Get the marketplace information for a given name.
---

# opennebula_marketplace

Use this data source to retrieve the marketplace information from it's name or tags.

## Example Usage

```hcl
data "opennebula_marketplace" "example" {
  name = "My_Marketplace"
}
```

## Argument Reference

* `id` - (Optional) ID of the marketplace.
* `name` - (Optional) The OpenNebula marketplace to retrieve information for.
* `tags` - (Optional) Tags associated to the marketplace.

## Attribute Reference

* `id` - ID of the marketplace.
* `name` - The OpenNebula marketplace name.
* `tags` - Tags of the marketplace (Key = Value).
