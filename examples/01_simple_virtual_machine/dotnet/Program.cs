using System.Collections.Generic;
using Pulumi;
using Pulumiverse.EsxiNative;
using Pulumiverse.EsxiNative.Inputs;

return await Deployment.RunAsync(() =>
{
    var vm = new VirtualMachine("vm-test", new VirtualMachineArgs
    {
        DiskStore = "nvme-ssd-datastore",
        NetworkInterfaces = new NetworkInterfaceArgs[]
        {
            new ()
            {
                VirtualNetwork = "default"
            }
        }
    });

    return new Dictionary<string, object?>
    {
        ["id"] = vm.Id,
        ["name"] = vm.Name,
        ["os"] = vm.Os,
    };
});
