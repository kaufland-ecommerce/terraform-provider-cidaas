# Terraform Provider for CIDAAS

![GoLang](https://img.shields.io/badge/golang-1.21-blue)
![GitHub](https://img.shields.io/github/license/real-digital/terraform-provider-cidaas)

Terraform provider to manage and read resources from cidaas instances.

Currently developed against cidaas version `3.9`

* [Usage documentation for the Provider can be found in the Terraform Registry](https://registry.terraform.io/providers/kaufland-ecommerce/cidaas/latest/docs)
* [Additional examples can be found in the
  `./examples` folder within this repository](https://github.com/kaufland-ecommerce/terraform-provider-cidaas/tree/main/examples).

## Development

If you are new to developing custom providers for terraform, you can start
from [this hashicorp tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).

This terraform provider is developed by [framework plugin](https://developer.hashicorp.com/terraform/plugin/framework).

In order to run terraform project locally, we need to instruct terraform to use the local version of this provider
instead of the published version from terraform repository. To do so, please create `~/.terraformrc` with the following
contents:

```
provider_installation {

  dev_overrides {
      "kaufland-ecommerce/cidaas" = "<GOBIN_PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Where `GOBIN_PATH` is the path for the `go/bin` dir. On linux it is `/home/USER/go/bin`

In order to make this project and available to terraform (with latest changes), you need to install it in `GOBIN` dir
by:

```
go install .
```

## Debugging:

We will use vscode. In `.vscode/launch.json` put the following contents:

```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Terraform Provider",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            // this assumes your workspace is the root of the repo
            "program": "${workspaceFolder}",
            "env": {},
            "args": [
                "-debug",
            ]
        }
    ]
}

```

Then from the right pane:

1. place the desired breakpoints
2. select `Run and Debug`
3. select `Debug Terrafrom Provider`
4. press `Start debugging`
5. in `debug console`, you will see `TF_REATTACH_PROVIDERS='{"registry...` (value clipped as too long). Use this env var
   before each terraform operation, e.g. `TF_REATTACH_PROVIDERS='{"registry... terraform plan`
