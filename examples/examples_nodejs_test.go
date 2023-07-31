package examples

import (
	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"path/filepath"
	"testing"
)

func Test01SimpleVirtualMachineTs(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "01_simple_virtual_machine", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test02ClonedVirtualMachineCompleteBuild(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "02_cloned_virtual_machine_complete_build", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test03ResourcePoolsAdditionalStorage(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "03_resource_pools_additional_storage", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test04TalosLinux(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "04_talos_linux", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test05CloudInitAndTemplates(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "05_cloud_init_and_templates", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test06OVFProperties(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "06_ovf_properties", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test07Networking(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "07_networking", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func Test08NetworkingCloudInit(t *testing.T) {
	test := getNodeJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			Dir: filepath.Join(getCwd(t), "08_networking_cloud_init", "nodejs"),
		})

	integration.ProgramTest(t, &test)
}

func getNodeJSBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"@edmondshtogu/pulumi-esxi-native",
		},
	})

	return baseJS
}
