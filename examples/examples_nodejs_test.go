package examples

import (
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"path/filepath"
	"testing"
)

func Test01SimpleVirtualMachineTs(t *testing.T) {
	testExample("01_simple_virtual_machine", t)
}

func Test02ClonedVirtualMachineCompleteBuild(t *testing.T) {
	testExample("02_cloned_virtual_machine_complete_build", t)
}

func Test03ResourcePoolsAdditionalStorage(t *testing.T) {
	testExample("03_resource_pools_additional_storage", t)
}

func Test04TalosLinux(t *testing.T) {
	testExample("04_talos_linux", t)
}

func Test05CloudInitAndTemplates(t *testing.T) {
	testExample("05_cloud_init_and_templates", t)
}

func Test06OVFProperties(t *testing.T) {
	testExample("06_ovf_properties", t)
}

func Test07Networking(t *testing.T) {
	testExample("07_networking", t)
}

func Test08NetworkingCloudInit(t *testing.T) {
	testExample("08_networking_cloud_init", t)
}

func testExample(name string, t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), name, "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func getNodeJSBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"@pulumiverse/esxi-native",
		},
	})

	return baseJS
}
