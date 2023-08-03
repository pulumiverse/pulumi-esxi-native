package esxi

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func TestVirtualMachineCreate(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
	}

	inputs := getBaseVMInputs()

	esxi := getESXi(t)
	id, result, err := VirtualMachineCreate(inputs, esxi)
	if err != nil {
		t.Skipf("Test failed with err: %s", err)
	}

	if id == "" {
		t.FailNow()
	}

	if len(result) == 0 {
		t.FailNow()
	}
}

func getBaseVMInputs() resource.PropertyMap {
	inputs := resource.PropertyMap{
		"bootDiskSize": {V: float64(16)},
		"bootDiskType": {V: "thin"},
		"bootFirmware": {V: "bios"},
		"diskStore":    {V: "nvme-ssd-datastore"},
		"memSize":      {V: float64(512)},
		"name":         {V: "vm-test-9967a16"},
		"networkInterfaces": {V: []resource.PropertyValue{
			{V: resource.PropertyMap{
				"virtualNetwork": {V: "default"},
			}},
		}},
		"numVCpus":           {V: float64(1)},
		"os":                 {V: "centos"},
		"ovfPropertiesTimer": {V: float64(6000)},
		"resourcePoolName":   {V: "/"},
		"shutdownTimeout":    {V: float64(600)},
		"startupTimeout":     {V: float64(600)},
		"virtualHWVer":       {V: float64(13)},
	}
	return inputs
}

func getESXi(t *testing.T) *Host {
	t.Helper()
	// Open the .env file
	file, err := os.Open(filepath.Join(getCwd(t), "../../../examples/.env"))
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

	var host, sshPort, sslPort, user, pass string

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
			host = value
		case "ESXI_USERNAME":
			user = value
		case "ESXI_PASSWORD":
			pass = value
		case "ESXI_SSH_PORT":
			sshPort = value
		case "ESXI_SSL_PORT":
			sslPort = value
		}
	}

	if err := scanner.Err(); err != nil {
		t.Skipf("Skipping test due failure on reading .env file! Err: %s", err)
		return nil
	}

	esxiHost, err := NewHost(host, sshPort, sslPort, user, pass)
	if err != nil {
		t.Skipf("Skipping test due failure on building ESXi host! Err: %s", err)
	}
	return esxiHost
}

func getCwd(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}
