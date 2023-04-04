---
layout: "opennebula"
page_title: "OpenNebula: opennebula_zone"
sidebar_current: "docs-opennebula-datasource-zone"
description: |-
  Get the zone information for a given name.
---

# opennebula_zone

Use this data source to retrieve the zone information from it's name.

## Example Usage

```hcl
data "opennebula_zone" "example" {
  name = "My_Zone"
}
```

## Argument Reference

* `id` - (Optional) ID of the zone.
* `name` - (Optional) The OpenNebula zone to retrieve information for.

## Attribute Reference

* `id` - ID of the zone.
* `name` - The OpenNebula zone name.
