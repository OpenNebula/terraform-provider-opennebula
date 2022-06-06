---
layout: "opennebula"
page_title: "OpenNebula: opennebula_user"
sidebar_current: "docs-opennebula-resource-user"
description: |-
  Provides an OpenNebula user resource.
---

# opennebula_user

Provides an OpenNebula user resource.

This resource allows you to manage users on your OpenNebula clusters. When applied,
a new user is created. When destroyed, it is removed.

## Example Usage

```hcl
resource "opennebula_user" "example" {
  name          = "user"
  password      = "randomp4ss"
  auth_driver   = "core"
  primary_group = "100"
  groups        = [101, 102]

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
    network_quotas {
      id     = 11
      leases = 4
    }
    image_quotas {
      id          = 8
      running_vms = 1
    }
    image_quotas {
      id          = 9
      running_vms = 1
    }
  }

  tags = {
    environment = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user.
* `password` - (Optional) Password of the user. It is required for all `auth_driver` excepted 'ldap'
* `auth_driver` - (Optional) Authentication Driver for User management. DEfaults to 'core'.
* `primary_group` - (Optional) Primary group ID of the User. Defaults to 0 (oneadmin).
* `groups` - (Optional) List of secondary groups ID of the user.
* `quotas` - (Optional) See [Quotas parameters](#quotas-parameters) below for details
* `tags` - (Optional) Group tags (Key = value)

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

* `id` - ID of the user.

## Import

`opennebula_user` can be imported using its ID:

```shell
terraform import opennebula_user.example 123
```
