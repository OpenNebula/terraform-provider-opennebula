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
data "opennebula_template" "example" {
  name = "My_Template"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula template to retrieve information for.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `context` - (Deprecated) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include: `NETWORK = "YES"` and `SET_HOSTNAME = "$NAME"`.
* `graphics` - (Deprecated) Graphics parameters.
* `os` - (Deprecated) OS parameters
* `tags` - (Optional) Template tags (Key = Value).

## Attribute Reference

The following attributes are exported:

* `id` - ID of the template.
* `name` - Name of the template.
* `cpu` - Amount of CPU shares assigned to the VM.
* `vpcu` - Number of CPU cores presented to the VM.
* `memory` - Amount of RAM assigned to the VM in MB.
* `disk` - Disk parameters
* `nic` - NIC parameters
* `vmgroup` - VM group parameters
* `tags` - Tags of the template (Key = Value).
