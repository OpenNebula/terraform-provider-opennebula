---
layout: "opennebula"
page_title: "OpenNebula: opennebula_hook"
sidebar_current: "docs-opennebula-datasource-hook"
description: |-
  Get information about an OpenNebula hook.
---

# opennebula_hook

Use this data source to retrieve information about an existing OpenNebula hook.

## Example Usage

```hcl
data "opennebula_hook" "example" {
  name = "log-new-user"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) ID of the hook.
* `name` - (Optional) Name of the hook.
* `tags` - (Optional) Map of tags used to filter the hook.

## Attribute Reference

The following attributes are exported:

* `name` - Name of the hook.
* `type` - Hook type: `api` or `state`.
* `command` - Command executed when the hook is triggered.
* `arguments` - Arguments passed to the hook command.
* `arguments_stdin` - Whether arguments are passed through stdin.
* `timeout` - Hook execution timeout in seconds.
* `call` - API call that triggers an API hook.
* `resource` - Resource type for state hooks.
* `remote` - Whether the state hook executes remotely.
* `state` - State that triggers a state hook.
* `lcm_state` - LCM state that triggers a VM state hook.
* `on` - VM state transition shortcut.
* `tags` - Custom hook template tags, excluding native hook attributes.
