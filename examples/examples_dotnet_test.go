//go:build dotnet || all
// +build dotnet all

package examples

import (
	"testing"
)

func Test01SimpleVirtualMachineCs(t *testing.T) {
	testExample("01_simple_virtual_machine", DOTNET, t)
}

func Test02ClonedVirtualMachineCompleteBuildCs(t *testing.T) {
	testExample("02_cloned_virtual_machine_complete_build", DOTNET, t)
}

func Test03ResourcePoolsAdditionalStorageCs(t *testing.T) {
	testExample("03_resource_pools_additional_storage", DOTNET, t)
}

func Test04TalosLinuxCs(t *testing.T) {
	testExample("04_talos_linux", DOTNET, t)
}

func Test05CloudInitAndTemplatesCs(t *testing.T) {
	testExample("05_cloud_init_and_templates", DOTNET, t)
}

func Test06OVFPropertiesCs(t *testing.T) {
	testExample("06_ovf_properties", DOTNET, t)
}

func Test07NetworkingCs(t *testing.T) {
	testExample("07_networking", DOTNET, t)
}

func Test08NetworkingCloudInitCs(t *testing.T) {
	testExample("08_networking_cloud_init", DOTNET, t)
}
