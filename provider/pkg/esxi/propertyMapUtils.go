package esxi

import "github.com/pulumi/pulumi/sdk/v3/go/common/resource"

func (vm *VirtualMachine) ToPropertyMap() resource.PropertyMap {
	outputs := resource.PropertyMap{
		"name": resource.PropertyValue{
			V: vm.Name,
		},
		"bootFirmware": resource.PropertyValue{
			V: vm.BootFirmware,
		},
		"diskStore": resource.PropertyValue{
			V: vm.DiskStore,
		},
		"resourcePoolName": resource.PropertyValue{
			V: vm.ResourcePoolName,
		},
		"bootDiskSize": resource.PropertyValue{
			V: vm.BootDiskSize,
		},
		"memSize": resource.PropertyValue{
			V: vm.MemSize,
		},
		"numVCpus": resource.PropertyValue{
			V: vm.NumVCpus,
		},
		"virtualHWVer": resource.PropertyValue{
			V: vm.VirtualDisks,
		},
		"os": resource.PropertyValue{
			V: vm.Os,
		},
		"power": resource.PropertyValue{
			V: vm.Power,
		},
		"ipAddress": resource.PropertyValue{
			V: vm.IpAddress,
		},
		"startupTimeout": resource.PropertyValue{
			V: vm.StartupTimeout,
		},
		"shutdownTimeout": resource.PropertyValue{
			V: vm.ShutdownTimeout,
		},
		"notes": resource.PropertyValue{
			V: vm.Notes,
		},
	}

	if vm.BootDiskType != "Unknown" && vm.BootDiskType != "" {
		outputs["bootDiskType"] = resource.PropertyValue{
			V: vm.BootDiskType,
		}
	}

	if len(vm.Info) > 0 {
		info := make([]resource.PropertyMap, len(vm.Info))
		for i, vmi := range vm.Info {
			info[i] = vmi.ToPropertyMap()
		}
		outputs["info"] = resource.PropertyValue{
			V: info,
		}
	}

	//if len(vm.Info) > 0 {
	//	outputs["info"] = resource.PropertyValue{
	//		V: vm.Info,
	//	}
	//}

	// Do network interfaces
	if len(vm.NetworkInterfaces) > 0 && vm.NetworkInterfaces[0].VirtualNetwork != "" {
		networkInterfaces := make([]resource.PropertyMap, len(vm.NetworkInterfaces))
		for i, ni := range vm.NetworkInterfaces {
			networkInterfaces[i] = ni.ToPropertyMap()
		}
		outputs["networkInterfaces"] = resource.PropertyValue{
			V: networkInterfaces,
		}
	}

	//if len(vm.NetworkInterfaces) > 0 && vm.NetworkInterfaces[0].VirtualNetwork != "" {
	//	outputs["networkInterfaces"] = resource.PropertyValue{
	//		V: vm.NetworkInterfaces,
	//	}
	//}

	// Do virtual disks
	if len(vm.VirtualDisks) > 0 && vm.VirtualDisks[0].VirtualDiskId != "" {
		virtualDisks := make([]resource.PropertyMap, len(vm.VirtualDisks))
		for i, vd := range vm.VirtualDisks {
			virtualDisks[i] = vd.ToPropertyMap()
		}
		outputs["virtualDisks"] = resource.PropertyValue{
			V: virtualDisks,
		}
	}

	//if len(vm.VirtualDisks) > 0 && vm.VirtualDisks[0].VirtualDiskId != "" {
	//	outputs["virtualDisks"] = resource.PropertyValue{
	//		V: vm.VirtualDisks,
	//	}
	//}

	return outputs
}

func (ni *KeyValuePair) ToPropertyMap() resource.PropertyMap {
	outputs := resource.PropertyMap{
		"key": resource.PropertyValue{
			V: ni.Key,
		},
		"value": resource.PropertyValue{
			V: ni.Value,
		},
	}

	return outputs
}

func (ni *Uplink) ToPropertyMap() resource.PropertyMap {
	outputs := resource.PropertyMap{
		"name": resource.PropertyValue{
			V: ni.Name,
		},
	}

	return outputs
}

func (ni *NetworkInterface) ToPropertyMap() resource.PropertyMap {
	outputs := resource.PropertyMap{
		"virtualNetwork": resource.PropertyValue{
			V: ni.VirtualNetwork,
		},
		"macAddress": resource.PropertyValue{
			V: ni.MacAddress,
		},
		"nicType": resource.PropertyValue{
			V: ni.NicType,
		},
	}

	return outputs
}

func (ni *VMVirtualDisk) ToPropertyMap() resource.PropertyMap {
	outputs := resource.PropertyMap{
		"virtualDiskId": resource.PropertyValue{
			V: ni.VirtualDiskId,
		},
		"slot": resource.PropertyValue{
			V: ni.Slot,
		},
	}

	return outputs
}
