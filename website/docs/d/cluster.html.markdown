---
layout: "opennebula"
page_title: "OpenNebula: opennebula_cluster"
sidebar_current: "docs-opennebula-datasource-cluster"
description: |-
  Get the cluster information for a given name.
---

# opennebula_cluster

Use this data source to retrieve the cluster information from it's name or tags.

## Example Usage

```hcl
data "opennebula_cluster" "example" {
  name = "My_Cluster"
}
```

## Argument Reference

* `id` - (Optional) ID of the cluster.
* `name` - (Optional) The OpenNebula cluster to retrieve information for.
* `tags` - (Optional) Tags associated to the cluster.

## Attribute Reference

* `id` - ID of the cluster.
* `name` - The OpenNebula cluster name.
* `tags` - Tags of the cluster (Key = Value).
