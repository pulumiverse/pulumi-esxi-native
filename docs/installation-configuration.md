---
title: ESXi Native Installation & Configuration
meta_desc: Information on how to install the Pulumi ESXi Native provider.
layout: package
---

## Installation

The Pulumi ESXi Native provider is available as a package in all Pulumi languages:

* JavaScript/TypeScript: [`@pulumiverse/esxi-native`](https://www.npmjs.com/package/@pulumiverse/esxi-native)
* Python: [`pulumiverse_esxi_native`](https://pypi.org/project/pulumiverse_esxi_native/)
* Go: [`github.com/pulumiverse/pulumi-esxi-native/sdk/go/esxi`](https://pkg.go.dev/github.com/pulumiverse/pulumi-esxi-native/sdk/go/esxi)
* .NET: [`Pulumiverse.EsxiNative`](https://www.nuget.org/packages/Pulumiverse.EsxiNative)

### Provider Binary

The ESXi Native provider binary is a third party binary. It can be installed using the `pulumi plugin` command.

```bash
pulumi plugin install resource esxi-native <version> --server github://api.github.com/pulumiverse
```

Replace the `<version>` string with your desired version.
