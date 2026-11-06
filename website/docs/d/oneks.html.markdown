---
layout: "opennebula"
page_title: "OpenNebula: opennebula_oneks"
sidebar_current: "docs-opennebula-datasource-oneks"
description: |-
  Looks up a Kubernetes cluster through the OneKS API.
---

# opennebula_oneks

Reads an existing Kubernetes cluster through the OneKS API. Configure the provider's
`oneks_endpoint` or `OPENNEBULA_ONEKS_ENDPOINT`, including the `/api/v1` prefix.

## Example Usage

```hcl
data "opennebula_oneks" "example" {
  name = "tf-oneks-demo"
}

output "cluster_state" {
  value = data.opennebula_oneks.example.state
}
```

## Argument Reference

Supply at least one selector:

* `id` - (Optional, Computed) Cluster ID. Zero is valid. Uses `GET /clusters/:id`.
* `name` - (Optional, Computed) Cluster name. Name-only lookup uses `GET /clusters` and must match exactly one cluster.

When both selectors are supplied, the fetched cluster must also match the name.

## Attribute Reference

* `id` - Cluster ID.
* `name` - Cluster name.
* `description` - Cluster description.
* `kubernetes_version` - Kubernetes version.
* `deployment` - Deployment placement as nested blocks, using the same structure as the resource. Access the cluster ID with `deployment[0].cluster[0].id`, public and private network IDs with `deployment[0].networks[0].public[0].id` and `deployment[0].networks[0].private[0].id`, and the optional image datastore ID with `deployment[0].datastores[0].image[0].id`.
* `state` - Current OneKS lifecycle state.
