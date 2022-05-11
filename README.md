[![CI](https://github.com/OpenNebula/terraform-provider-opennebula/actions/workflows/ci.yaml/badge.svg)](https://github.com/OpenNebula/terraform-provider-opennebula/actions/workflows/ci.yaml)

<a href="https://terraform.io">
    <img src="https://raw.githubusercontent.com/hashicorp/terraform-website/master/public/img/logo-text.svg" alt="Terraform logo" title="Terraform" height="30" />
</a> &nbsp; <a href="https://opennebula.io/">
    <img src="https://opennebula.io/wp-content/uploads/2013/12/opennebula_cloud_logo_white_bg.png" alt="OpenNebula logo" title="OKTA" height="30" />
</a>

# Terraform Provider for OpenNebula

## Quick Start

The documentation is available in the [Terraform Registry](https://registry.terraform.io/providers/OpenNebula/opennebula/latest/docs). There are lot of examples and a complete reference there.

## Contribute

[Bug Reports](https://github.com/OpenNebula/terraform-provider-opennebula/issues/new?template=bug.md), [Feature Requests](https://github.com/OpenNebula/terraform-provider-opennebula/issues/new?template=feature.md) and [Pull Requests](https://github.com/OpenNebula/terraform-provider-opennebula/compare) are welcome. Please follow [How to Contribute](https://github.com/OpenNebula/one/wiki/How-to-participate-in-Add_on-Development) rules for any Pull Request.

### Team

- Tino Vazquez ([tinova](https://github.com/tinova))
- François Rousselet ([frousselet](https://github.com/frousselet))
- Jean-Philippe Fourès ([jaypif](https://github.com/jaypif))
- Pierre Lafièvre ([treywelsh](https://github.com/treywelsh))
- Edouard Hur ([hekmon](https://github.com/hekmon))
- Benjamin Gustin ([aloababa](https://github.com/aloababa))

## Compatibility

* Leverages [OpenNebula's XML/RPC API](https://docs.opennebula.io/5.12/integration/system_interfaces/api.html)
* Tested on OpenNebula version 5.12

This provider has been initiated to use official Goca from [OpenNebula](https://github.com/OpenNebula/one)

For older OpenNebula and Terraform releases, you can use non official provider from [Runtastic](https://github.com/runtastic/terraform-provider-opennebula) and enhanced by [BlackBerry](https://github.com/blackberry/terraform-provider-opennebula).

## Requirements

### Terraform

Because this Add-On is the OpenNebula Terraform Provider, it requires to have Terraform installed on your machine.
Instructions to install terraform are accessible [here](https://learn.hashicorp.com/terraform/getting-started/install)

Please note that this version is indended to be used with Terraform version 0.12+

### Golang (for testing)

OpenNebula Terraform provider is written in Golang, you must have a Golang environment to compile it.

A Golang dependency management tool is also required. This README is based on [goland/dep](https://github.com/golang/dep)

## Build from sources

1. Get the code of the OpenNebula provider
2. Get provider dependencies (if you use go dep)

```shell
$ dep init
```
3. Compile

```shell
$ go build -o terraform-provider-opennebula
```

**Warning: this provider is a "Third party" provider. It must follow these rules for the binary name.**

## References

Other Projects about Terraform provider exists. This project has been inspired by [Runtastic](https://github.com/runtastic/terraform-provider-opennebula) and [BlackBerry](https://github.com/blackberry/terraform-provider-opennebula) projects

## License

This project is under MPL v2.0 License. For more details about the License, please read LICENCE file.
