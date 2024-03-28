---
layout: "opennebula"
page_title: "OpenNebula: opennebula_templates"
sidebar_current: "docs-opennebula-datasource-templates"
description: |-
  Get the template information for a given name.
---

# opennebula_templates

Use this data source to retrieve templates information.

## Example Usage

```hcl
data "opennebula_templates" "example" {
  name_regex = "test.*"
  has_cpu    = true
  sort_on    = "register_date"
  order      = "ASC"
}
```


## Argument Reference

* `name_regex` - (Optional) Filter templates by name with a RE2 regular expression.
* `sort_on` - (Optional) Attribute used to sort the template list among: `id`, `name`, `cpu`, `vcpu`, `memory`, `register_date`.
* `has_cpu` - (Optional) Indicate if a CPU value has been defined.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `has_vcpu` - (Optional) Indicate if a VCPU value has been defined.
* `vcpu` - (Optional) Number of CPU cores presented to the VM.
* `has_memory` - (Optional) Indicate if a memory value has been defined.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `tags` - (Optional) Template tags (Key = Value).
* `order` - (Optional) Ordering of the sort: ASC or DESC.

## Attribute Reference

The following attributes are exported:

* `templates` - For each filtered template, this section collect a list of attributes. See [templates attributes](#templates-attributes)

## Templates attributes

* `id` - ID of the template.
* `name` - Name of the template.
* `cpu` - Amount of CPU shares assigned to the VM.
* `vcpu` - Number of CPU cores presented to the VM.
* `memory` - Amount of RAM assigned to the VM in MB.
* `disk` - Disk parameters
* `nic` - NIC parameters
* `vmgroup` - VM group parameters
* `register_date` - Creation date of the template
* `tags` - Tags of the template (Key = Value).
