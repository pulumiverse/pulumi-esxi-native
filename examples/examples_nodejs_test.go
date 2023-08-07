//go:build nodejs || all
// +build nodejs all

package examples

import (
	"testing"
)

func Test01SimpleVirtualMachineTs(t *testing.T) {
	testExample("01_simple_virtual_machine", NODEJS, t)
}

func Test02ClonedVirtualMachineCompleteBuildTs(t *testing.T) {
	testExample("02_cloned_virtual_machine_complete_build", NODEJS, t)
}

func Test03ResourcePoolsAdditionalStorageTs(t *testing.T) {
	testExample("03_resource_pools_additional_storage", NODEJS, t)
}

func Test04TalosLinuxTs(t *testing.T) {
	testExample("04_talos_linux", NODEJS, t)
}

func Test05CloudInitAndTemplatesTs(t *testing.T) {
	testExample("05_cloud_init_and_templates", NODEJS, t)
}

func Test06OVFPropertiesTs(t *testing.T) {
	testExample("06_ovf_properties", NODEJS, t)
}

func Test07NetworkingTs(t *testing.T) {
	testExample("07_networking", NODEJS, t)
}

func Test08NetworkingCloudInitTs(t *testing.T) {
	testExample("08_networking_cloud_init", NODEJS, t)
}
