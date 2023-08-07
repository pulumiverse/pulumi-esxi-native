//go:build python || all
// +build python all

package examples

import (
	"testing"
)

func Test01SimpleVirtualMachinePy(t *testing.T) {
	testExample("01_simple_virtual_machine", PYTHON, t)
}

func Test02ClonedVirtualMachineCompleteBuildPy(t *testing.T) {
	testExample("02_cloned_virtual_machine_complete_build", PYTHON, t)
}

func Test03ResourcePoolsAdditionalStoragePy(t *testing.T) {
	testExample("03_resource_pools_additional_storage", PYTHON, t)
}

func Test04TalosLinuxPy(t *testing.T) {
	testExample("04_talos_linux", PYTHON, t)
}

func Test05CloudInitAndTemplatesPy(t *testing.T) {
	testExample("05_cloud_init_and_templates", PYTHON, t)
}

func Test06OVFPropertiesPy(t *testing.T) {
	testExample("06_ovf_properties", PYTHON, t)
}

func Test07NetworkingPy(t *testing.T) {
	testExample("07_networking", PYTHON, t)
}

func Test08NetworkingCloudInitPy(t *testing.T) {
	testExample("08_networking_cloud_init", PYTHON, t)
}
