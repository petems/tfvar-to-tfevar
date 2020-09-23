# `tfvar-to-tfevar`

**tfvar-to-tfevar** is a way to export [Terraform](https://www.terraform.io/)'s [variable definitions](https://www.terraform.io/docs/configuration/variables.html#assigning-values-to-root-module-variables) to [Terraform Enteprise or Terraform Cloud variables](https://www.terraform.io/docs/cloud/workspaces/variables.html)

It is heavily based on the awesome original work in [tfvar](https://github.com/shihanng/tfvar)

For Terraform configuration that has input variables declared, e.g.,

```terraform
variable "image_id" {
  type = string
}

variable "availability_zone_names" {
  type    = list(string)
  default = ["us-west-1a"]
}

variable "docker_ports" {
  type = list(object({
    internal = number
    external = number
    protocol = string
  }))
  default = [
    {
      internal = 8300
      external = 8300
      protocol = "tcp"
    }
  ]
}
```

It will create valid Terraform code for the TFE Provider as individual `tfe_variable` resources.

```
$ tfvar-to-tfevar .
resource "tfe_variable" "availability_zone_names" {
  key          = "availability_zone_names"
  value        = <<EOT
availability_zone_names = ["us-west-1a"]
EOT
  category     = "terraform"
  hcl          = true
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""
}

resource "tfe_variable" "docker_ports" {
  key          = "docker_ports"
  value        = <<EOT
docker_ports = [{
  external = 8300
  internal = 8300
  protocol = "tcp"
}]
EOT
  category     = "terraform"
  hcl          = true
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""
}

resource "tfe_variable" "image_id" {
  key          = "image_id"
  value        = <<EOT
image_id = null
EOT
  category     = "terraform"
  hcl          = true
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""
}
```

If you then add a valid `tfvars` file, it'll read that as well and set values for the inputs:

```
$ tfvar-to-tfevar . --var-file='test.tfvars'
resource "tfe_variable" "availability_zone_names" {
  key          = "availability_zone_names"
  value        = <<EOT
availability_zone_names = ["us-west-1a"]
EOT
  category     = "terraform"
  hcl          = true
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""
}

resource "tfe_variable" "docker_ports" {
  key          = "docker_ports"
  value        = <<EOT
docker_ports = [{
  external = 8300
  internal = 8300
  protocol = "tcp"
}]
EOT
  category     = "terraform"
  hcl          = true
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""
}

resource "tfe_variable" "image_id" {
  key          = "image_id"
  value        = "xyz"
  category     = "terraform"
  workspace_id = data.tfe_workspace.example_workspace.id
  description  = ""

```

## Installation

### [Homebrew (macOS)](https://github.com/shihanng/homebrew-tfvar)

**WIP**

### Debian, Ubuntu


### Binaries

The [release page](https://github.com/petems/tfvar-to-tfevar/releases) contains binaries built for various platforms. Download the version matches your environment (e.g. `linux_amd64`) and place the binary in the executable `$PATH` e.g. `/usr/local/bin`:

```
curl -sL https://github.com/petems/tfvar-to-tfevar/releases/latest/download/tfvar_linux_amd64.tar.gz | \
    tar xz -C /usr/local/bin/ tfvar
```

### For Gophers

With [Go](https://golang.org/doc/install) already installed in your system, use `go get`

```
go get github.com/petems/tfvar-to-tfevar
```

or clone this repo and `make install`

```
git clone https://github.com/petems/tfvar-to-tfevar.git
cd tfvar
make install
```

## Contributing

Want to add missing feature? Found bug :bug:? Pull requests and issues are welcome. For major changes, please open an issue first to discuss what you would like to change :heart:.

```
make lint
make test
```

should help with the idiomatic Go styles and unit-tests.

## License
[MIT](./LICENSE)
