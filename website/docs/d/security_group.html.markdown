---
layout: "opennebula"
page_title: "OpenNebula: opennebula_security_group"
sidebar_current: "docs-opennebula-datasource-security-group"
description: |-
  Get the security group information for a given name.
---

# opennebula_security_group

Use this data source to retrieve the security group information for a given name.

## Example Usage

```hcl
data "opennebula_security_group" "example" {
  name = "My_Security_Group"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula security group to retrieve information for.
* `tags` - (Optional) Security group tags.

## Attribute Reference

* `id` - ID of the security group.
* `name` - Name of the security group
* `tags` - Tags of the security group (Key = Value).
