using Pulumi;
using Pulumi.EsxiNative;
using Pulumi.EsxiNative.Inputs;

class MyStack : Stack
{
    public MyStack()
    {
        var vm = new VirtualMachine("vm-test", new VirtualMachineArgs
        {
            DiskStore = "nvme-ssd-datastore",
            NetworkInterfaces = new NetworkInterfaceArgs[]
            {
                {
                    VirtualNetwork = "default"
                }
            }
        });

        this.Id = vm.Id;
        this.Name = vm.Name;
        this.Os = vm.Os;
    }

    [Output]
    public Output<string> Id { get; set; }
    [Output]
    public Output<string> Name { get; set; }
    [Output]
    public Output<string> Os { get; set; }
}
