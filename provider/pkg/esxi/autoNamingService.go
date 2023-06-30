package esxi

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

type AutoNamingSpec struct {
	AutoName  string
	MinLength int
	MaxLength int
}

type nameSpec struct {
	minLength int
	maxLength int
}
type AutoNamingService struct {
	rules map[string]nameSpec
}

func NewAutoNamingService() *AutoNamingService {
	return &AutoNamingService{
		rules: map[string]nameSpec{
			"esxi-native:index:PortGroup":      {3, 250},
			"esxi-native:index:ResourcePool":   {5, 250},
			"esxi-native:index:VirtualDisk":    {3, 250},
			"esxi-native:index:VirtualMachine": {5, 250},
			"esxi-native:index:VirtualSwitch":  {3, 250},
		},
	}
}

func (service *AutoNamingService) CreateAutoNamingSpec(token string, inputProperties resource.PropertyMap) *AutoNamingSpec {
	// AutoNaming
	defaultNameSpec, ok := service.rules[token]
	if !ok {
		return nil
	}

	var autoNameSpec *AutoNamingSpec

	if propSpec, has := inputProperties["name"]; has && propSpec.IsString() {
		autoNameSpec = &AutoNamingSpec{
			AutoName:  "name",
			MinLength: defaultNameSpec.minLength,
			MaxLength: defaultNameSpec.maxLength,
		}
	}

	return autoNameSpec
}
