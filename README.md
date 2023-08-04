# Pulumi Native ESXi Provider

This repository crates a VMWare ESXi provider to provision VMs directly on an ESXi hypervisor without a need for vCenter or vSphere.

[![ci](https://github.com/pulumiverse/pulumi-esxi-native/actions/workflows/ci.yaml/badge.svg)](https://github.com/pulumiverse/pulumi-esxi-native/actions/workflows/ci.yaml) [![release](https://github.com/pulumiverse/pulumi-esxi-native/actions/workflows/release.yml/badge.svg)](https://github.com/pulumiverse/pulumi-esxi-native/actions/workflows/release.yml)

The repository is created based on the Terraform Provider [terraform-provider-esxi](https://github.com/josenk/terraform-provider-esxi/tree/master).
Thanks to the wonderful work done there by [@josenk](https://github.com/josenk), I was able to build this provider for Pulumi users.

## IMPORTANT NOTES!

**Note for Pull Requests (PRs)**: When creating a pull request, please do it onto the **MAIN branch** which is
the consolidated work-in-progress branch. Do not request it onto another branch.

> **PLEASE** Read our [branch guide](branch-guide.md) to know about our branching policy
>
> ### CONTRIBUTING
>
> **IMPORTANT:** The contribution details are stated [here](CONTRIBUTING.md)

## Requirements
-   The VMware [ovftool](https://customerconnect.vmware.com/downloads/get-download?downloadGroup=OVFTOOL443) is required to be installed in the workstation where pulumi will be executed.  
    > **NOTE:** `ovftool` installer for windows doesn't put ovftool.exe in your path. 
      You will need to manually set your path.
-   You MUST enable ssh access on your ESXi hypervisor.
    > Google 'How to enable ssh access on esxi'
      >- In general, you should know how to use terraform, esxi and some networking...
* You will most likely need a DHCP server on your primary network if you are deploying VMs with public OVF/OVA/VMX images.  (Sources that have unconfigured primary interfaces.)
- The source OVF/OVA/VMX images must have open-vm-tools or vmware-tools installed to properly import an IPAddress.  (you need this to run provisioners)

## Features and Compatibility

* Source image can be a clone of a VM or local vmx, ovf, ova file. This provider uses ovftool, so there should be a wide compatibility.
* Supports adding your VM to Resource Pools to partition CPU and memory usage from other VMs on your ESXi host.
* Pulumi will Create, Destroy, Update & Import Resource Pools.
* Pulumi will Create, Destroy, Update & Import Virtual Machines.
* Pulumi will Create, Destroy, Update & Import Virtual Disks.
* Pulumi will Create, Destroy, Update & Import Virtual Switches.
* Pulumi will Create, Destroy, Update & Import Port Groups.

## Why this provider?

If you do not have a vCenter or vSphere, especially if you are running a home lab, these services are expensive, and maybe you cannot have them, but the ESXi is free, so that is the reason behind it!

## How to install

TODO: Details will be added here

## How to use and configure

TODO: Details will be added here

## Available resources

TODO: Details will be added here

## Known issues with vmware_esxi

* Using a local source vmx files should not have any networks configured. There is very limited network interface mapping abilities in packer for vmx files.  
  It's best to simply clean out all network information from your vmx file. The plugin will add network configuration to the destination vm guest as required.
* pulumi import cannot import the guest disk type (thick, thin, etc.) if the VM is powered on and cannot import the guest `ipAddress` if it's powered off.
* Only `numVCpus` are supported, `numCores` is not.
* Doesn't support CD-ROM or floppy.
* Doesn't support Shared bus Interfaces, or Shared disks.
* Using an incorrect password could lockout your account using default esxi pam settings.
* Don't set `startupTimeout` or `shutdownTimeout` to 0 (zero). It's valid, however it will be changed to default values.

