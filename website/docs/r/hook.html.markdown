---
layout: "opennebula"
page_title: "OpenNebula: opennebula_hook"
sidebar_current: "docs-opennebula-resource-hook"
description: |-
  Provides an OpenNebula hook resource.
---

# opennebula_hook

Provides an OpenNebula hook resource.

Hooks execute commands in response to OpenNebula API calls or resource state transitions.

## Example Usage

API hook:

```hcl
resource "opennebula_hook" "api" {
  name            = "log-new-user"
  type            = "api"
  command         = "log_new_user.rb"
  arguments       = "$API"
  arguments_stdin = true
  call            = "one.user.allocate"
}
```

VM state hook:

```hcl
resource "opennebula_hook" "vm_pending" {
  name      = "vm-pending"
  type      = "state"
  command   = "vm-pending.rb"
  arguments = "$TEMPLATE pending"
  resource  = "VM"
  on        = "CUSTOM"
  state     = "PENDING"
  lcm_state = "LCM_INIT"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the hook.
* `type` - (Required) Hook type. Valid values are `api` and `state`.
* `command` - (Required) Command executed when the hook is triggered. Relative paths are resolved by OpenNebula's hook execution manager.
* `arguments` - (Optional) Arguments passed to the hook command.
* `arguments_stdin` - (Optional) Pass arguments through stdin instead of command line arguments.
* `timeout` - (Optional) Hook execution timeout in seconds. If omitted, OpenNebula uses an infinite timeout.
* `call` - (Required for `type = "api"`) API call that triggers the hook, for example `one.user.allocate`.
* `resource` - (Required for `type = "state"`) Resource type for state hooks. Valid values are `IMAGE`, `HOST`, and `VM`.
* `remote` - (Optional) Execute the state hook remotely on the host that triggered the hook or runs the VM.
* `state` - (Required for `HOST`/`IMAGE` state hooks and for `VM` hooks with `on = "CUSTOM"`) State that triggers the hook.
* `lcm_state` - (Required for `VM` state hooks with `on = "CUSTOM"`) LCM state that triggers the hook.
* `on` - (Required for `VM` state hooks) Shortcut for common VM state transitions. Valid values are `PROLOG`, `RUNNING`, `SHUTDOWN`, `STOP`, `DONE`, `UNKNOWN`, and `CUSTOM`.
* `tags` - (Optional) Map of custom tags to add to the hook template. Hook template attributes such as `CALL`, `COMMAND`, `STATE`, and `ON` cannot be used as tag keys.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the hook.
* `default_tags` - Default tags applied from provider configuration.
* `tags_all` - Combined map of resource tags and provider default tags.

## Import

`opennebula_hook` can be imported using its ID:

```shell
terraform import opennebula_hook.example 123
```
