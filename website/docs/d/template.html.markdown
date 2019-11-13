---
layout: "opennebula"
page_title: "OpenNebula: opennebula_template"
sidebar_current: "docs-opennebula-datasource-template"
description: |-
  Get the template information for a given name.
---

# opennebula_template

Use this data source to retrieve the template information for a given name.

## Example Usage

```hcl
data "opennebula_template" "ExistingTemplate" {
  name = "My_Template"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula template to retrieve information for.

