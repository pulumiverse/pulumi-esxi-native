// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export class ResourcePool extends pulumi.CustomResource {
    /**
     * Get an existing ResourcePool resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): ResourcePool {
        return new ResourcePool(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'esxi-native:index:ResourcePool';

    /**
     * Returns true if the given object is an instance of ResourcePool.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is ResourcePool {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === ResourcePool.__pulumiType;
    }

    /**
     * CPU maximum (in MHz).
     */
    public readonly cpuMax!: pulumi.Output<number | undefined>;
    /**
     * CPU minimum (in MHz).
     */
    public readonly cpuMin!: pulumi.Output<number | undefined>;
    /**
     * Can pool borrow CPU resources from parent?
     */
    public readonly cpuMinExpandable!: pulumi.Output<string | undefined>;
    /**
     * CPU shares (low/normal/high/<custom>).
     */
    public readonly cpuShares!: pulumi.Output<string | undefined>;
    /**
     * Memory maximum (in MB).
     */
    public readonly memMax!: pulumi.Output<number | undefined>;
    /**
     * Memory minimum (in MB).
     */
    public readonly memMin!: pulumi.Output<number | undefined>;
    /**
     * Can pool borrow memory resources from parent?
     */
    public readonly memMinExpandable!: pulumi.Output<string | undefined>;
    /**
     * Memory shares (low/normal/high/<custom>).
     */
    public readonly memShares!: pulumi.Output<string | undefined>;
    /**
     * Resource Pool Name
     */
    public readonly name!: pulumi.Output<string>;

    /**
     * Create a ResourcePool resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args?: ResourcePoolArgs, opts?: pulumi.CustomResourceOptions) {
        let resourceInputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            resourceInputs["cpuMax"] = args ? args.cpuMax : undefined;
            resourceInputs["cpuMin"] = (args ? args.cpuMin : undefined) ?? 100;
            resourceInputs["cpuMinExpandable"] = (args ? args.cpuMinExpandable : undefined) ?? "true";
            resourceInputs["cpuShares"] = (args ? args.cpuShares : undefined) ?? "normal";
            resourceInputs["memMax"] = args ? args.memMax : undefined;
            resourceInputs["memMin"] = (args ? args.memMin : undefined) ?? 200;
            resourceInputs["memMinExpandable"] = (args ? args.memMinExpandable : undefined) ?? "true";
            resourceInputs["memShares"] = (args ? args.memShares : undefined) ?? "normal";
            resourceInputs["name"] = args ? args.name : undefined;
        } else {
            resourceInputs["cpuMax"] = undefined /*out*/;
            resourceInputs["cpuMin"] = undefined /*out*/;
            resourceInputs["cpuMinExpandable"] = undefined /*out*/;
            resourceInputs["cpuShares"] = undefined /*out*/;
            resourceInputs["memMax"] = undefined /*out*/;
            resourceInputs["memMin"] = undefined /*out*/;
            resourceInputs["memMinExpandable"] = undefined /*out*/;
            resourceInputs["memShares"] = undefined /*out*/;
            resourceInputs["name"] = undefined /*out*/;
        }
        opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts);
        super(ResourcePool.__pulumiType, name, resourceInputs, opts);
    }
}

/**
 * The set of arguments for constructing a ResourcePool resource.
 */
export interface ResourcePoolArgs {
    /**
     * CPU maximum (in MHz).
     */
    cpuMax?: pulumi.Input<number>;
    /**
     * CPU minimum (in MHz).
     */
    cpuMin?: pulumi.Input<number>;
    /**
     * Can pool borrow CPU resources from parent?
     */
    cpuMinExpandable?: pulumi.Input<string>;
    /**
     * CPU shares (low/normal/high/<custom>).
     */
    cpuShares?: pulumi.Input<string>;
    /**
     * Memory maximum (in MB).
     */
    memMax?: pulumi.Input<number>;
    /**
     * Memory minimum (in MB).
     */
    memMin?: pulumi.Input<number>;
    /**
     * Can pool borrow memory resources from parent?
     */
    memMinExpandable?: pulumi.Input<string>;
    /**
     * Memory shares (low/normal/high/<custom>).
     */
    memShares?: pulumi.Input<string>;
    /**
     * Resource Pool Name
     */
    name?: pulumi.Input<string>;
}
