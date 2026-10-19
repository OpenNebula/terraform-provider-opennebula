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

```hcl
terraform {
  required_providers {
    opennebula = {
      source = "OpenNebula/opennebula"
      version = "~> 1.4"
    }
  }
}

provider "opennebula" {
  endpoint      = "https://example.com:2633/RPC2"
}

resource "opennebula_group" "group" {
  name = "OpenNebula"
}
```

## Authentication and Configuration

The configuration of the OpenNebula Provider can be set by the `provider` block attributes or by the environment variables.

### Provider attributes

* `endpoint` - (Required) The URL of OpenNebula XML-RPC Endpoint API (for example, `http://example.com:2633/RPC2`).
* `flow_endpoint` - (Optional) The OneFlow HTTP Endpoint API (for example, `http://example.com:2474`).
* `username` - (Required) The OpenNebula username.
* `password` - (Required) The Opennebula password matching the username.
* `insecure` - (Optional) Allow insecure connexion (skip TLS verification).
* `default_tags` - (Optional) Apply default custom tags to resources supporting `tags`. Theses tags can be overriden in the `tags` section of the resource. See [Using tags](#using-tags) below for more details.

!> **Warning:** Hard-coded credentials are not recommended in any Terraform configuration file and should not be commited in a public repository you might prefer [Environment variables instead](#environment-variables).

### Environment variables

The provider can also read the following environment variables if no value is set in the the `provider` block attributes:

* `OPENNEBULA_ENDPOINT`
* `OPENNEBULA_FLOW_ENDPOINT`
* `OPENNEBULA_USERNAME`
* `OPENNEBULA_PASSWORD`
* `OPENNEBULA_INSECURE`

### Example

```bash
export OPENNEBULA_ENDPOINT="https://example.com:2633/RPC2"
export OPENNEBULA_FLOW_ENDPOINT="https://example.com:2474"
export OPENNEBULA_USERNAME="me"
export OPENNEBULA_PASSWORD="p@s5w0rD"
export OPENNEBULA_INSECURE="true"
```

```hcl
provider "opennebula" {}

resource "opennebula_group" "group" {
  # ...
}
```

```bash
terraform init
terraform plan
```

## Proxy support

Connections to both the OpenNebula XML-RPC API and OneFlow respect the standard proxy environment variables:

* `HTTP_PROXY` / `http_proxy` - Proxy URL for HTTP endpoints.
* `HTTPS_PROXY` / `https_proxy` - Proxy URL for HTTPS endpoints. An HTTP proxy uses CONNECT to tunnel HTTPS connections.
* `NO_PROXY` / `no_proxy` - Comma-separated hosts, domains, IP addresses, or CIDR ranges that bypass the proxy. Use `*` to bypass it for all hosts.

Uppercase variables take precedence over lowercase variables. Without a matching proxy variable, connections are direct. Localhost and loopback addresses always bypass the proxy. TLS verification remains enabled unless `insecure` is set.

Set these variables before starting Terraform or OpenTofu:

```bash
export HTTP_PROXY="http://proxy.example.com:3128"
export HTTPS_PROXY="http://proxy.example.com:3128"
export NO_PROXY="localhost,127.0.0.1,.internal.example.com"
tofu plan
```

No additional provider attributes are required. Proxy URLs may include authentication credentials; treat these as secrets and do not commit them or share them in logs. An HTTP endpoint sends API traffic without TLS, including OpenNebula authentication data, through the proxy. Prefer HTTPS endpoints when using a proxy.

## Using tags

### Resource tags

Some resources can support the `tags` attribute. In consists of a map of key-value elements. In the OpenNebula language, these are called 'attributes'.

```hcl
tags = {
  environment = "production"
}
```

### Default tags

The provider's `default_tags` attribute allows to set default tags for all resources supporting tags.

`default_tags` supports the following arguments:

* `tags` - (Optional) Map of tags.

When a tag is added to a resource, it overrides the one present in `default_tags` from the provider block if the key matches. In the example bellow, the resource will have `environment = "production"` and `deployment_method = "terraform"` tags.

```hcl
provider "opennebula" {
  endpoint      = "https://example.com:2633/RPC2"

  default_tags {
    tags = {
      environment     = "default"
      deployment_mode = "terraform"
    }
  }
}

resource "opennebula_group" "group" {
  # ...

  tags = {
    environment = "production"
  }
}
```
