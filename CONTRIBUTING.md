# Contributing to the Terraform Provider for OpenNebula

Thanks for getting involved in the Terraform Provider for OpenNebula. Here are a few step to read before contributing to this project.

## Set up you development environment

### Tools

* [Install Go](https://go.dev/doc/install)
* [Install Terraform](https://learn.hashicorp.com/terraform/getting-started/install)

### Local development of the provider

To develop and use this provider locally, you can leverage th [`dev_override`](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) feature of the terraform CLI.
To do so, create a `$HOME/.terraformrc` file if it doesn't exist yet (see the [`terraform` CLI documentation](https://developer.hashicorp.com/terraform/cli/config/config-file#locations) for different consideration for Windows systems) with the following content:

```hcl
provider_installation {
  dev_overrides {
    "OpenNebula/opennebula" = "[LOCAL_PATH_TO_THIS_REPO]"
  }
  direct {}
}
```

This configuration will not apply any change for `terraform init` or your `.terraform.lock.hcl`. From the Terraform documentation:
```text
With development overrides in effect, the terraform init command will still attempt to select a suitable published
version of your provider to install and record in the dependency lock file for future use, but other commands like
terraform apply will disregard the lock file's entry and will use the given directory instead.
```

This configuration will check for the `terraform-provider-opennebula` binary at the given directory. You can generate it using `make build`

Example:
```hcl
terraform {
  required_providers {
    opennebula = {
      source = "OpenNebula/opennebula" # use the real provider as source
      version = "1.4.0"
    }
  }
}

provider "opennebula" {
  # ...
}

data opennebula_datastore "my_datastore" {
  # ...
}
```

Applying changes for the above simple terraform module shows the override in action:
```txt
$> terraform apply
╷
│ Warning: Provider development overrides are in effect
│ 
│ The following provider development overrides are set in the CLI configuration:
│  - opennebula/opennebula in [LOCAL_PATH_TO_THIS_REPO]
│ 
│ The behavior may therefore not match any released version of the provider and applying changes may cause the state to become incompatible with published releases.
╵
data.opennebula_datastore.my_datastore: Reading...
data.opennebula_datastore.my_datastore: Read complete after 0s

No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration and found no differences, so no changes are needed.

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
```

## Issues and Pull Requests

You must use existing templates for Issues and Pull Requests.

### Issues

Please follow the following rules:

* Please vote on the issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment mentioning this issue

### Pull Request

Please follow the following rules:

* Please vote on the Pull Request by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original Pull Request to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for PR followers and do not help prioritize the request

If you are intereseted working on an issue, open a new Pull Request as _Draft_ and start working on it.

### Stale Issues and Pull Requests

Issues and Pull Requests are automaticaly labeled as `stale` after 30 days. Without any action, the Issue or the Pull request is closed 5 days after.

### Quality

A Pull Request must satisfy the following requierements:

* I have created an issue and I have mentioned it in `References`
* My code follows the style guidelines of this project (use `go fmt`)
* My changes generate no new warnings or errors
* I have updated the unit tests and they pass succesfuly
* I have commented my code, particularly in hard-to-understand areas
* I have updated the documentation (if needed)
* I have updated the changelog file

## Need help?

Please contact us. We will be glad to help you.
