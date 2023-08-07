package examples

import (
	"bufio"
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

func testExample(name string, sdk SDK, t *testing.T) {
	var opts integration.ProgramTestOptions

	switch sdk {
	case DOTNET:
		opts = getDotnetBaseOptions(t).
			With(integration.ProgramTestOptions{
				Dir: filepath.Join(getCwd(t), name, "dotnet"),
			})
	case NODEJS:
		opts = getNodeJSBaseOptions(t).
			With(integration.ProgramTestOptions{
				Dir: filepath.Join(getCwd(t), name, "nodejs"),
			})
	case GO:
		opts = getGoBaseOptions(t).
			With(integration.ProgramTestOptions{
				Dir: filepath.Join(getCwd(t), name, "go"),
			})
	case PYTHON:
		opts = getPythonBaseOptions(t).
			With(integration.ProgramTestOptions{
				Dir: filepath.Join(getCwd(t), name, "python"),
			})
	}

	integration.ProgramTest(t, &opts)
}

func getDotnetBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseCsharp := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"Pulumiverse.EsxiNative",
		},
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
	config := getConfig(t)

	return integration.ProgramTestOptions{
		Config:               config,
		ExpectRefreshChanges: true,
		SkipRefresh:          true,
		Quick:                true,
	}
}

func getConfig(t *testing.T) map[string]string {
	config := make(map[string]string)

	// Open the .env file
	file, err := os.Open(filepath.Join(getCwd(t), ".env"))
	if err != nil {
		t.Skipf("Skipping test due failure on reading .env file! Err: %s", err)
		return nil
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
			config["esxi-native:config:host"] = value
		case "ESXI_USERNAME":
			config["esxi-native:config:username"] = value
		case "ESXI_PASSWORD":
			config["esxi-native:config:password"] = value
		case "ESXI_SSH_PORT":
			config["esxi-native:config:sshPort"] = value
		case "ESXI_SSL_PORT":
			config["esxi-native:config:sslPort"] = value
		}
	}

	if err := scanner.Err(); err != nil {
		t.Skipf("Skipping test due failure on reading .env file! Err: %s", err)
		return nil
	}

	return config
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}
