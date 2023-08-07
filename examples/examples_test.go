package examples

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

type SDK string

const (
	DOTNET SDK = "dotnet"
	NODEJS SDK = "nodejs"
	GO     SDK = "go"
	PYTHON SDK = "python"
)

type Test struct {
	Name string
	Path string
	SDK  SDK
}

func TestExamples(t *testing.T) {
	for _, test := range getTests(t) {
		t.Run(test.Name, func(t *testing.T) {
			var opts integration.ProgramTestOptions

			switch test.SDK {
			case DOTNET:
				opts = getDotnetBaseOptions(t)
			case NODEJS:
				opts = getNodeJSBaseOptions(t)
			case GO:
				opts = getGoBaseOptions(t)
			case PYTHON:
				opts = getPythonBaseOptions(t)
			}
			opts = opts.
				With(integration.ProgramTestOptions{
					Dir: test.Path,
				})

			integration.ProgramTest(t, &opts)
		})
	}
}

func getDotnetBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseCsharp := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"Pulumiverse.EsxiNative",
		},
		Env: []string{fmt.Sprintf("PULUMI_LOCAL_NUGET='%s'", filepath.Join(getCwd(t), "../nuget"))},
	})

	return baseCsharp
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

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseGo := base.With(integration.ProgramTestOptions{
		Verbose: true,
		Dependencies: []string{
			"github.com/pulumiverse/pulumi-esxi-native/sdk",
		},
	})

	return baseGo
}

func getPythonBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	basePy := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			filepath.Join("..", "sdk", "python", "bin"),
		},
	})

	return basePy
}

func getBaseOptions(t *testing.T) integration.ProgramTestOptions {
	configs, secrets := getConfigsAndSecrets(t)

	return integration.ProgramTestOptions{
		Config:               configs,
		Secrets:              secrets,
		ExpectRefreshChanges: true,
		SkipRefresh:          true,
		Quick:                true,
	}
}

func getConfigsAndSecrets(t *testing.T) (map[string]string, map[string]string) {
	configs := make(map[string]string)
	secrets := make(map[string]string)

	// Open the .env file
	file, err := os.Open(filepath.Join(getCwd(t), ".env"))
	if err != nil {
		t.Skipf("Skipping test due failure on reading .env file! Err: %s", err)
		return nil, nil
	}
	defer func(file *os.File) {
		e := file.Close()
		if e != nil {
			log.Fatal(e)
		}
	}(file)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract the key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Println("Invalid line:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ESXI_HOST":
			configs["esxi-native:config:host"] = value
		case "ESXI_USERNAME":
			configs["esxi-native:config:username"] = value
		case "ESXI_PASSWORD":
			secrets["esxi-native:config:password"] = value
		case "ESXI_SSH_PORT":
			configs["esxi-native:config:sshPort"] = value
		case "ESXI_SSL_PORT":
			configs["esxi-native:config:sslPort"] = value
		}
	}

	if err = scanner.Err(); err != nil {
		t.Skipf("Skipping test due failure on reading .env file! Err: %s", err)
		return nil, nil
	}

	return configs, secrets
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}

func getTests(t *testing.T) []Test {
	var tests []Test

	baseDir := getCwd(t)
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		sdk := SDK(info.Name())

		if sdk == GO || sdk == DOTNET || sdk == NODEJS || sdk == PYTHON {
			test := Test{
				Name: strings.ReplaceAll(path, "/", "_"),
				Path: filepath.Join(baseDir, path),
				SDK:  sdk,
			}

			tests = append(tests, test)
		}

		return nil
	})

	if err != nil {
		t.Errorf("error walking directory: %v", err)
		return make([]Test, 0)
	}

	return tests
}
