---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_machines"
sidebar_current: "docs-opennebula-datasource-virtual-machines"
description: |-
  Get the virtual machine information for a given name.
---

# opennebula_virtual_machines

Use this data source to retrieve virtual machines information.

## Example Usage

```hcl
data "opennebula_virtual_machines" "example" {
  name_regex = "test.*"
  sort_on    = "id"
  order      = "ASC"
}
```


## Argument Reference

* `name_regex` - (Optional) Filter virtual machines by name with a RE2 regular expression.
* `sort_on` - (Optional) Attribute used to sort the VMs list among: `id`, `name`, `cpu`, `vcpu`, `memory`.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `vcpu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `tags` - (Optional) virtual machine tags (Key = Value).
* `order` - (Optional) Ordering of the sort: ASC or DESC.

## Attribute Reference

The following attributes are exported:

* `virtual_machines` - For each filtered virtual machine, this section collect a list of attributes. See [virtual-machines-attributes](#virtual-machines-attributes)

## Virtual machines attributes

* `id` - ID of the virtual machine.
* `name` - Name of the virtual machine.
* `cpu` - Amount of CPU shares assigned to the VM.
* `vcpu` - Number of CPU cores presented to the VM.
* `memory` - Amount of RAM assigned to the VM in MB.
* `disk` - Disk parameters.
* `nic` - NIC parameters.
* `nic_alias` - NIC Alias parameters.
* `vmgroup` - VM group parameters
* `tags` - Tags of the virtual machine (Key = Value).
