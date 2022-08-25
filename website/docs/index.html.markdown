---
layout: "opennebula"
page_title: "Provider: OpenNebula"
sidebar_current: "docs-opennebula-index"
description: |-
  The OpenNebula provider is used to interact with OpenNebula cluster resources.
---

# OpenNebula Provider

The OpenNebula provider is used to interact with OpenNebula cluster resources.

The provider allows you to manage your OpenNebula clusters resources.
It needs to be configured with proper credentials before it can be used.

## Example Usage

Configure the OpenNebula Provider:

```hcl
provider "opennebula" {
  endpoint      = "<ENDPOINT URL>"
  flow_endpoint = "<FLOW ENDPOINT URL>"
  username      = "<USERNAME>"
  password      = "<PASSWORD OR TOKEN>"
}
```

Create a new group of users to the OpenNebula cluster:

```hcl
resource "opennebula_group" "group" {
  # ...
}
```

## Argument Reference

The following arguments are mandatory in the `provider` block:

* `endpoint` - (Required) The URL of OpenNebula XML-RPC Endpoint API (for example, `http://example.com:2633/RPC2`).
* `flow_endpoint` - (Optional) The OneFlow HTTP Endpoint API (for example, `http://example.com:2474/RPC2`).
* `username` - (Required) The OpenNebula Username.
* `password` - (Required) The Opennebula Password of the username.
* `insecure` - (Optional) Allow insecure connexion (skip TLS verification).
