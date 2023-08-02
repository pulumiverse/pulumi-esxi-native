package esxi

type AutoNamingSpec struct {
	PropertyName string
	MinLength    int
	MaxLength    int
}

type AutoNamingService struct {
	rules map[string]AutoNamingSpec
}

func NewAutoNamingService() *AutoNamingService {
	return &AutoNamingService{
		rules: map[string]AutoNamingSpec{
			"esxi-native:index:PortGroup":      {"name", 3, 250},
			"esxi-native:index:ResourcePool":   {"name", 5, 250},
			"esxi-native:index:VirtualDisk":    {"name", 3, 250},
			"esxi-native:index:VirtualMachine": {"name", 5, 250},
			"esxi-native:index:VirtualSwitch":  {"name", 3, 250},
		},
	}
}

func (service *AutoNamingService) GetAutoNamingSpec(token string) *AutoNamingSpec {
	// AutoNaming
	autoNameSpec, ok := service.rules[token]
	if !ok {
		return nil
	}

	return &autoNameSpec
}
