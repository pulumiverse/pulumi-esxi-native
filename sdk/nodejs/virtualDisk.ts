// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import { input as inputs, output as outputs, enums } from "./types";
import * as utilities from "./utilities";

export class VirtualDisk extends pulumi.CustomResource {
    /**
     * Get an existing VirtualDisk resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): VirtualDisk {
        return new VirtualDisk(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'esxi-native:index:VirtualDisk';

    /**
     * Returns true if the given object is an instance of VirtualDisk.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is VirtualDisk {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === VirtualDisk.__pulumiType;
    }

    /**
     * Disk directory.
     */
    public readonly directory!: pulumi.Output<string>;
    /**
     * Disk Store.
     */
    public readonly diskStore!: pulumi.Output<string>;
    /**
     * Virtual Disk type. (thin, zeroedthick or eagerzeroedthick)
     */
    public readonly diskType!: pulumi.Output<enums.DiskType>;
    /**
     * Virtual Disk Name.
     */
    public readonly name!: pulumi.Output<string>;
    /**
     * Virtual Disk size in GB.
     */
    public readonly size!: pulumi.Output<number | undefined>;

    /**
     * Create a VirtualDisk resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: VirtualDiskArgs, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.directory === undefined) && !opts.urn) {
                throw new Error("Missing required property 'directory'");
            }
            if ((!args || args.diskStore === undefined) && !opts.urn) {
                throw new Error("Missing required property 'diskStore'");
            }
            if ((!args || args.diskType === undefined) && !opts.urn) {
                throw new Error("Missing required property 'diskType'");
            }
            resourceInputs["directory"] = args ? args.directory : undefined;
            resourceInputs["diskStore"] = args ? args.diskStore : undefined;
            resourceInputs["diskType"] = args ? args.diskType : undefined;
            resourceInputs["name"] = args ? args.name : undefined;
            resourceInputs["size"] = (args ? args.size : undefined) ?? 1;
        } else {
            resourceInputs["directory"] = undefined /*out*/;
            resourceInputs["diskStore"] = undefined /*out*/;
            resourceInputs["diskType"] = undefined /*out*/;
            resourceInputs["name"] = undefined /*out*/;
            resourceInputs["size"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(VirtualDisk.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a VirtualDisk resource.
 */
export interface VirtualDiskArgs {
    /**
     * Disk directory.
     */
    directory: pulumi.Input<string>;
    /**
     * Disk Store.
     */
    diskStore: pulumi.Input<string>;
    /**
     * Virtual Disk type. (thin, zeroedthick or eagerzeroedthick)
     */
    diskType: pulumi.Input<enums.DiskType>;
    /**
     * Virtual Disk Name.
     */
    name?: pulumi.Input<string>;
    /**
     * Virtual Disk size in GB.
     */
    size?: pulumi.Input<number>;
}
