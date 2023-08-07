//go:build go || all
// +build go all

package examples

import (
	"testing"
)

func Test01SimpleVirtualMachineGo(t *testing.T) {
	testExample("01_simple_virtual_machine", GO, t)
}

func Test02ClonedVirtualMachineCompleteBuildGo(t *testing.T) {
	testExample("02_cloned_virtual_machine_complete_build", GO, t)
}

func Test03ResourcePoolsAdditionalStorageGo(t *testing.T) {
	testExample("03_resource_pools_additional_storage", GO, t)
}

func Test04TalosLinuxGo(t *testing.T) {
	testExample("04_talos_linux", GO, t)
}

func Test05CloudInitAndTemplatesGo(t *testing.T) {
	testExample("05_cloud_init_and_templates", GO, t)
}

func Test06OVFPropertiesGo(t *testing.T) {
	testExample("06_ovf_properties", GO, t)
}

func Test07NetworkingGo(t *testing.T) {
	testExample("07_networking", GO, t)
}

func Test08NetworkingCloudInitGo(t *testing.T) {
	testExample("08_networking_cloud_init", GO, t)
}
