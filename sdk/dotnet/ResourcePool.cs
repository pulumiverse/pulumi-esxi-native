// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;
using Pulumi;

namespace Pulumiverse.EsxiNative
{
    [EsxiNativeResourceType("esxi-native:index:ResourcePool")]
    public partial class ResourcePool : global::Pulumi.CustomResource
    {
        /// <summary>
        /// CPU maximum (in MHz).
        /// </summary>
        [Output("cpuMax")]
        public Output<int?> CpuMax { get; private set; } = null!;

        /// <summary>
        /// CPU minimum (in MHz).
        /// </summary>
        [Output("cpuMin")]
        public Output<int?> CpuMin { get; private set; } = null!;

        /// <summary>
        /// Can pool borrow CPU resources from parent?
        /// </summary>
        [Output("cpuMinExpandable")]
        public Output<string?> CpuMinExpandable { get; private set; } = null!;

        /// <summary>
        /// CPU shares (low/normal/high/&lt;custom&gt;).
        /// </summary>
        [Output("cpuShares")]
        public Output<string?> CpuShares { get; private set; } = null!;

        /// <summary>
        /// Memory maximum (in MB).
        /// </summary>
        [Output("memMax")]
        public Output<int?> MemMax { get; private set; } = null!;

        /// <summary>
        /// Memory minimum (in MB).
        /// </summary>
        [Output("memMin")]
        public Output<int?> MemMin { get; private set; } = null!;

        /// <summary>
        /// Can pool borrow memory resources from parent?
        /// </summary>
        [Output("memMinExpandable")]
        public Output<string?> MemMinExpandable { get; private set; } = null!;

        /// <summary>
        /// Memory shares (low/normal/high/&lt;custom&gt;).
        /// </summary>
        [Output("memShares")]
        public Output<string?> MemShares { get; private set; } = null!;

        /// <summary>
        /// Resource Pool Name
        /// </summary>
        [Output("name")]
        public Output<string> Name { get; private set; } = null!;


        /// <summary>
        /// Create a ResourcePool resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public ResourcePool(string name, ResourcePoolArgs? args = null, CustomResourceOptions? options = null)
            : base("esxi-native:index:ResourcePool", name, args ?? new ResourcePoolArgs(), MakeResourceOptions(options, ""))
        {
        }

        private ResourcePool(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("esxi-native:index:ResourcePool", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
                PluginDownloadURL = "github://api.github.com/pulumiverse",
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing ResourcePool resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static ResourcePool Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new ResourcePool(name, id, options);
        }
    }

    public sealed class ResourcePoolArgs : global::Pulumi.ResourceArgs
    {
        /// <summary>
        /// CPU maximum (in MHz).
        /// </summary>
        [Input("cpuMax")]
        public Input<int>? CpuMax { get; set; }

        /// <summary>
        /// CPU minimum (in MHz).
        /// </summary>
        [Input("cpuMin")]
        public Input<int>? CpuMin { get; set; }

        /// <summary>
        /// Can pool borrow CPU resources from parent?
        /// </summary>
        [Input("cpuMinExpandable")]
        public Input<string>? CpuMinExpandable { get; set; }

        /// <summary>
        /// CPU shares (low/normal/high/&lt;custom&gt;).
        /// </summary>
        [Input("cpuShares")]
        public Input<string>? CpuShares { get; set; }

        /// <summary>
        /// Memory maximum (in MB).
        /// </summary>
        [Input("memMax")]
        public Input<int>? MemMax { get; set; }

        /// <summary>
        /// Memory minimum (in MB).
        /// </summary>
        [Input("memMin")]
        public Input<int>? MemMin { get; set; }

        /// <summary>
        /// Can pool borrow memory resources from parent?
        /// </summary>
        [Input("memMinExpandable")]
        public Input<string>? MemMinExpandable { get; set; }

        /// <summary>
        /// Memory shares (low/normal/high/&lt;custom&gt;).
        /// </summary>
        [Input("memShares")]
        public Input<string>? MemShares { get; set; }

        /// <summary>
        /// Resource Pool Name
        /// </summary>
        [Input("name")]
        public Input<string>? Name { get; set; }

        public ResourcePoolArgs()
        {
        }
        public static new ResourcePoolArgs Empty => new ResourcePoolArgs();
    }
}
