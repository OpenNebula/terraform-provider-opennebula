# Contributing to the Terraform Provider for OpenNebula

Thanks for getting involved in the Terraform Provider for OpenNebula. Here are a few step to read before contributing to this project.

## Set up you development environment

### Tools

* [Install Go](https://go.dev/doc/install)
* [Install Terraform](https://learn.hashicorp.com/terraform/getting-started/install)

### Building from sources

```shell
export tf_arch=darwin_arm64
export tf_one_version=0.0.1

# Clone terraform-provider-opennebula
git clone git@github.com:OpenNebula/terraform-provider-opennebula.git

# Create directory under Terraform plugins directory
mkdir -p ${HOME}/.terraform.d/plugins/one.test/one/opennebula/${tf_one_version}/${tf_arch}

# Create a link to the Provider binary
ln -s $(pwd)/terraform-provider-opennebula/terraform-provider-opennebula ${HOME}/.terraform.d/plugins/one.test/one/opennebula/${tf_one_version}/${tf_arch}

# Build the Provider
cd terraform-provider-opennebula
go build
```

Now you can create a new `main.tf` file:

```hcl
terraform {
  required_providers {
    opennebula = {
      source  = "one.test/one/opennebula"
    }
  }
}

provider "opennebula" {
  # ...
}

resource "opennebula_image" "image" {
  # ...
}
```

During the `terraform init`, the provider should be initialized as `unauthenticated`:

```text
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding latest version of one.test/one/opennebula...
- Installing one.test/one/opennebula v0.0.1...
- Installed one.test/one/opennebula v0.0.1 (unauthenticated)

Terraform has created a lock file .terraform.lock.hcl to record the provider
selections it made above. Include this file in your version control repository
so that Terraform can guarantee to make the same selections by default when
you run "terraform init" in the future.

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
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

Issues and Pull Requests are automaticaly labeled as `stale` after 55 days. Without any action, the Issue or the Pull request is closed 5 days after.

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
