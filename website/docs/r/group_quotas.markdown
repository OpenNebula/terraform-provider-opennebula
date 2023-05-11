---
layout: "opennebula"
page_title: "OpenNebula: opennebula_group_quotas"
sidebar_current: "docs-opennebula-resource-group"
description: |-
  Provides an OpenNebula group resource.
---

# opennebula_group_quotas

Provides an OpenNebula group quotas resource.

This resource allows you to manage the quotas for a group.


## Example Usage

```hcl
resource "opennebula_group_quotas" "example" {
    group_id = opennebula_group.example.id
    datastore {
      id     = 1
      images = 5
      size   = 10000
    }
    vm{
      cpu            = 3
      running_cpu    = 3
      memory         = 2048
      running_memory = 2048
    }
    network {
      id     = 10
      leases = 6
    }
    network {
      id     = 11
      leases = 4
    }
    image {
      id          = 8
      running_vms = 1
    }
    image {
      id          = 9
      running_vms = 1
    }
}
```

## Argument Reference

The following arguments are supported:

* `group_id` - (Required) The related group ID.
* `datastore` - (Optional) List of datastore quotas. See [Datastore quotas parameters](#datastore-quotas-parameters) below for details.
* `network` - (Optional) List of network quotas. See [Network quotas parameters](#network-quotas-parameters) below for details.
* `image` - (Optional) List of image quotas. See [Image quotas parameters](#image-quotas-parameters) below for details
* `vm` - (Optional) See [Virtual Machine quotas parameters](#virtual-machine-quotas-parameters) below for details

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

## Import

`opennebula_group_quotas` can be imported using the group ID and the quota section to import:

```shell
terraform import opennebula_group_quotas.example 123:image
```
