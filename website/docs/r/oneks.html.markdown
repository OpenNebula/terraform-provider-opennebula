---
layout: "opennebula"
page_title: "OpenNebula: opennebula_oneks"
sidebar_current: "docs-opennebula-resource-oneks"
description: |-
  Manages a Kubernetes cluster through the OneKS lifecycle API.
---

# opennebula_oneks

Creates and provisions a Kubernetes cluster through the OneKS API. Creation waits
for the cluster to reach `RUNNING`. Destruction requests deprovisioning and waits
until the API confirms that the cluster has been deleted.

## Example Usage

```hcl
provider "opennebula" {
  endpoint       = "https://opennebula.example.com:2633/RPC2"
  oneks_endpoint = "https://oneks.example.com/api/v1"
}

resource "opennebula_oneks" "example" {
  name               = "tf-oneks-demo"
  kubernetes_version = "v1.35.0"

  deployment {
    cluster { id = 0 }
    networks {
      public  { id = 1 }
      private { id = 2 }
    }
    # Optional image datastore placement:
    # datastores {
    #   image { id = 3 }
    # }
  }

  spec {
    family  = "general"
    flavour = "standalone"
  }

  timeouts {
    create = "60m"
    delete = "60m"
  }
}
```

Choose a version, control-plane family, flavour, and input values supported by your
OneKS installation. The IDs in `deployment` identify existing OpenNebula placement
resources. OneKS validates their availability before creating the cluster.

## Argument Reference

* `name` - (Required) Cluster name, lowercase up to 63 characters.
* `kubernetes_version` - (Required) Kubernetes version supported by the control-plane family. Changing this argument recreates the cluster.
* `deployment` - (Required) A single placement block containing `cluster`, `networks.public`, and `networks.private` blocks, each with a nonnegative integer `id`. May also contain a `datastores` block with an `image` block and its `id`. Zero is a valid ID. Changes recreate the cluster.
* `spec` - (Required, Sensitive) A single control-plane creation block. Requires `flavour`; accepts `family` and optional JSON-encoded `user_inputs_values`. OneKS supplies family and input defaults. Changes recreate the cluster.
* `description` - (Optional, Computed) Nonempty description, updated through the OneKS API. Omitting it preserves the current description; the API does not support clearing it with an empty string.
* `force_delete` - (Optional) Request forced cleanup of a failed lifecycle operation during destruction. Defaults to `false`.

The provider requires `oneks_endpoint` or `OPENNEBULA_ONEKS_ENDPOINT`, including the
`/api/v1` prefix. It uses the provider's `username`, `password`, and `insecure` settings.

`spec` is a creation input retained in Terraform state. OneKS consumes it during
creation and stores the resolved control plane separately, so refresh does not
reconstruct the original input overrides or detect external changes to them.
Sensitive values in `spec` are hidden from normal plan output but are still stored
in Terraform state.

## User Input Overrides

The optional `spec.user_inputs_values` argument accepts a JSON object string.
Use `jsonencode` to supply mixed and nested HCL values:

```hcl
spec {
  family  = "general"
  flavour = "standalone"

  user_inputs_values = jsonencode({
    cpu    = 2
    vcpu   = 2
    memory = 4096
  })
}
```

Use names, types, and overrides supported by the selected family and flavour.
The provider sends the decoded object to OneKS, preserving
numbers, booleans, lists, objects, and null values. The argument is sensitive and
must encode an object; arrays, scalar values, and null at the top level are rejected.
Omit the argument to use the input defaults supplied by OneKS.

Changes to input values recreate the cluster. JSON formatting and object key order
do not cause changes. Refresh preserves the original creation inputs; import
reconstructs them as a JSON object string from the expanded control plane.

## Attribute Reference

* `id` - OneKS cluster ID.
* `state` - Current lifecycle state reported by OneKS.

## Timeouts and Failures

`create` and `delete` each default to 60 minutes. A lifecycle failure or timeout
returns an error and retains the allocated cluster ID. Inspect the cluster and its
logs through OneKS and recover failed operations before retrying. If deletion is
already in progress, a repeated destroy waits for that operation to finish.
Authentication errors, malformed responses, and API failures leave the ID in state.
Only HTTP 404 confirms that a cluster is absent.

## Import

```shell
terraform import opennebula_oneks.example 123
```

Import reads the expanded control plane through OneKS and reconstructs `spec` from
its `family`, `flavour`, and `user_inputs_values`, including resolved defaults.
Access to sensitive input values requires the permissions enforced by OneKS.
Match the imported `spec`, deployment, and Kubernetes version in your configuration
before applying to avoid replacement. Inspect the plan before applying.
