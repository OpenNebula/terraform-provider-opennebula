---
layout: "opennebula"
page_title: "OpenNebula: opennebula_group"
sidebar_current: "docs-opennebula-resource-group"
description: |-
  Provides an OpenNebula group resource.
---

# opennebula_group

Provides an OpenNebula group resource.

This resource allows you to manage groups on your OpenNebula clusters. When applied,
a new group is created. When destroyed, it is removed.

## Example Usage

```hcl
data "template_file" "grptpl" {
  template = file("group_template.txt")
}

resource "opennebula_group" "group" {
    name                  = "terraform"
    template              = data.template_file.grptpl.rendered
    delete_on_destruction = true
    quotas {
        datastore_quotas {
            id     = 1
            images = 3
            size   = 10000
        }
        vm_quotas {
            cpu            = 3
            running_cpu    = 3
            memory         = 2048
            running_memory = 2048
        }
        network_quotas = {
            id     = 10
            leases = 6
        }
        network_quotas = {
            id     = 11
            leases = 4
        }
        image_quotas = {
            id          = 8
            running_vms = 1
        }
        image_quotas = {
            id          = 9
            running_vms = 1
        }
    }
}
```

with `group_template.txt` file with Sunstone information:

```php
SUNSTONE = [
  DEFAULT_VIEW = "cloud",
  GROUP_ADMIN_DEFAULT_VIEW = "groupadmin",
  GROUP_ADMIN_VIEWS = "cloud,groupadmin",
  VIEWS = "cloud"
]
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the group.
* `template` - (Required) Group template content in OpenNebula XML or String format. Used to provide SUSNTONE arguments.
* `delete_on_destruction` - (Optional) Flag to delete the group on destruction. Defaults to `false`.
* `admins` - (Optional) List of Administrator user IDs part of the group.
* `quotas` - (Optional) See [Quotas parameters](#quotas-parameters) below for details

### Quotas parameters

`quotas` supports the following arguments:

* `datastore_quotas` - (Optional) List of datastore quotas. See [Datastore quotas parameters](#datastore-quotas-parameters) below for details.
* `network_quotas` - (Optional) List of network quotas. See [Network quotas parameters](#network-quotas-parameters) below for details.
* `image_quotas` - (Optional) List of image quotas. See [Image quotas parameters](#image-quotas-parameters) below for details
* `vm_quotas` - (Optional) See [Virtual Machine quotas parameters](#virtual-machine-quotas-parameters) below for details

#### Datastore quotas parameters

`datastore` supports the following arguments:

* `id` - (Required) Datastore ID.
* `images` - (Optional) Maximum number of images allowed on the datastore. Defaults to `default quota`
* `size` - (Optional) Total size in MB allowed on the datastore. Defaults to `default quota`

#### Network quotas parameters

`network` supports the following arguments:

* `id` - (Required) Network ID.
* `leases` - (Optional) Maximum number of ip leases allowed on the network. Defaults to `default quota`

#### Image quotas parameters

`image` supports the following arguments:

* `id` - (Required) Image ID.
* `running_vms` - (Optional) Maximum number of Virtual Machines in `RUNNING` state with this image ID attached. Defaults to `default quota`

#### Virtual Machine quotas parameters

`vm` supports the following arguments:

* `cpu` - (Optional) Total of CPUs allowed. Defaults to `default quota`.
* `memory` - (Optional) Total of memory (in MB) allowed. Defaults to `default quota`.
* `vms` - (Optional) Maximum number of Virtual Machines allowed. Defaults to `default quota`.
* `running_cpu` - (Optional) Virtual Machine CPUs allowed in `RUNNING` state. Defaults to `default quota`.
* `running_memory` - (Optional) Virtual Machine Memory (in MB) allowed in `RUNNING` state. Defaults to `default quota`.
* `running_vms` - (Optional) Number of Virtual Machines allowed in `RUNNING` state. Defaults to `default quota`.
* `system_disk_size` - (Optional) Maximum disk global size (in MB) allowed on a `SYSTEM` datastore. Defaults to `default quota`.

## Attribute Reference

The following attribute is exported:
* `id` - ID of the group.

## Import

To import an existing group #134 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_group" "importgroup" {
    name = "importedgroup"
}
```

And then run:

```
terraform import opennebula_group.importgroup 134
```

Verify that Terraform does not perform any change:

```
terraform plan
```
