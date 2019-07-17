# addon-terraform

## Description

[OpenNebula](https://opennebula.org/) provider for [Terraform](https://www.terraform.io/).

## Development

Bug reports and pull requests are welcome on GitHub at
https://github.com/OpenNebula/addon-terraform/issues.

Please follow [How to Contribute](https://github.com/OpenNebula/one/wiki/How-to-participate-in-Add_on-Development) rules for any pull request.

## Authors

* Leaders:

- Tino Vazquez (https://github.com/tinova)
- Jean-Philippe Fourès (https://github.com/jaypif)
- Pierre Lafièvre (https://github.com/treywelsh)
- Edouard Hur (https://github.com/hekmon)
- Benjamin Gustin (https://github.com/aloababa)

## Compatibility

* Leverages [OpenNebula's XML/RPC API](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html)
* Tested on OpenNebula version 5.8

This provider has been initiated to use official Goca from [OpenNebula](https://github.com/OpenNebula/one)

For Older OpenNebula and Terraform releases, you can use non official provider from [Runtastic](https://github.com/runtastic/terraform-provider-opennebula) and enhanced by [BlackBerry](https://github.com/blackberry/terraform-provider-opennebula).

## Features

### Data sources

Current definition of these data sources are supported:
* Groups
* Image
* Security Groups
* Template
* Virtual Data Center
* Virtual Network

### Resources

Current definition of these resources are supported:
* Groups [onegroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onegroup)
* Image [oneimage](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneimage)
* Security Groups [onesecgroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onesecgroup)
* Template [onetemplate](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onetemplate)
* Virtual Data Center [onevdc](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevdc)
* Virtual Machine [onevm](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevm)
* Virtual Network [onevnet](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevnet)

## Limitations

Following OpenNebula Objects **are not** currently supported:
* ACL [oneacl](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneacl)
* Accounting [oneacct](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneacct)
* Hosts Management [onehost](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onehost)
* Clusters [onecluster](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onecluster)
* Users [oneuser](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneuser)
* Datastore [onedatastore](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onedatastore)
* Market [onemarket](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onemarket)
* Market App [onemarketapp](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onemarketapp)
* Virtual Router [onevrouter](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevrouter)
* Zone [onezone](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onezone)

## Requirements

### Terraform

Because this Add-On is the OpenNebula Terraform Provider, it requires to have Terraform installed on your machine.
Instructions to install terraform are accessible [here](https://learn.hashicorp.com/terraform/getting-started/install)

Please note that this version is indended to be used with Terraform version 0.12+

### Golang

The only way to use this add-on is to compile it from source code.
OpenNebula Terraform provider is written in Golang, you must have a Golang environment to compile the provider.

A Golang dependency management tool is also required. This README is based on [goland/dep](https://github.com/golang/dep)

## Installation

### From Source

#### Compilation

1. Get the code of the OpenNebula provider.
2. Get provider dependencies (if you use go dep)
```
repopath$ dep init
```
3. Compile
```
repopath$ go build -o terraform-provider-opennebula
```

**Warning: this provider is a "Third party" provider. It must follow these rules for the binary name.**

#### Integration with Terraform

Create a terraform file to use OpenNebula provider (follow instructions on Wiki page of the project) and run `terraform init`.
This will initialize terraform to use OpenNebula Provider.

### With Terraform

*Work In Progress*

## Configuration

**Opennebula** provider has the following supports parameters:

| **Parameter** | **Description**                       |
| --------- | --------------------------------- |
| **endpoint**  | URL to the OpenNebula XML-RPC API |
| **username**  | OpenNebula username               |
| **password**  | OpenNebula password OR token      |
| **version**   | Version of the provider (optional) |

## Usage

Lots of Examples and details of data sources and resources parameters are available on the [Wiki](https://github.com/OpenNebula/addon-terraform/wiki).

## References

Other Projects about Terraform provider exists. This project has been inspired by [Runtastic](https://github.com/runtastic/terraform-provider-opennebula) and [BlackBerry](https://github.com/blackberry/terraform-provider-opennebula) projects

## License

This project is under MPL v2.0 License. For more details about the License, please read LICENCE file.
