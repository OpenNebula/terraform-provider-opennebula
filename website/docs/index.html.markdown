---
layout: "opennebula"
page_title: "Provider: OpenNebula"
sidebar_current: "docs-opennebula-index"
description: |-
  The OpenNebula provider is used to interact with OpenNebula cluster resources.
---

# OpenNebula Provider

The OpenNebula provider is used to interact with OpenNebula cluster resources.

The provider allow you to manage your OpenNebula clusters resources.
It needs to be configured with proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "opennebula" {
  endpoint      = "<ENDPOINT URL>"
  flow_endpoint = "<FLOW ENDPOINT URL>"
  username      = "<USERNAME>"
  password      = "<PASSWORD OR TOKEN>"
}

# Create a new group of users to the OpenNebula cluster
resource "opennebula_group" "group" {
  # ...
}
```

## Argument Reference

The following arguments are mandatory in the `provider` block:

* `endpoint` - (Required) This is the URL of OpenNebula XML-RPC Endpoint API (for example, `http://example.com:2633/RPC2`).
* `flow_endpoint` - (Optional) This is the OneFlow HTTP Endpoint API (for example, `http://example.com:2474/RPC2`).
* `username` - (Required) This is the OpenNebula Username.
* `password` - (Required) This is the Opennebula Password of the username.
