# coding=utf-8
# *** WARNING: this file was generated by pulumigen. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities

__all__ = ['ResourcePoolArgs', 'ResourcePool']

@pulumi.input_type
class ResourcePoolArgs:
    def __init__(__self__, *,
                 cpu_max: Optional[pulumi.Input[int]] = None,
                 cpu_min: Optional[pulumi.Input[int]] = None,
                 cpu_min_expandable: Optional[pulumi.Input[str]] = None,
                 cpu_shares: Optional[pulumi.Input[str]] = None,
                 mem_max: Optional[pulumi.Input[int]] = None,
                 mem_min: Optional[pulumi.Input[int]] = None,
                 mem_min_expandable: Optional[pulumi.Input[str]] = None,
                 mem_shares: Optional[pulumi.Input[str]] = None,
                 name: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a ResourcePool resource.
        :param pulumi.Input[int] cpu_max: CPU maximum (in MHz).
        :param pulumi.Input[int] cpu_min: CPU minimum (in MHz).
        :param pulumi.Input[str] cpu_min_expandable: Can pool borrow CPU resources from parent?
        :param pulumi.Input[str] cpu_shares: CPU shares (low/normal/high/<custom>).
        :param pulumi.Input[int] mem_max: Memory maximum (in MB).
        :param pulumi.Input[int] mem_min: Memory minimum (in MB).
        :param pulumi.Input[str] mem_min_expandable: Can pool borrow memory resources from parent?
        :param pulumi.Input[str] mem_shares: Memory shares (low/normal/high/<custom>).
        :param pulumi.Input[str] name: Resource Pool Name
        """
        if cpu_max is not None:
            pulumi.set(__self__, "cpu_max", cpu_max)
        if cpu_min is None:
            cpu_min = 100
        if cpu_min is not None:
            pulumi.set(__self__, "cpu_min", cpu_min)
        if cpu_min_expandable is None:
            cpu_min_expandable = 'true'
        if cpu_min_expandable is not None:
            pulumi.set(__self__, "cpu_min_expandable", cpu_min_expandable)
        if cpu_shares is None:
            cpu_shares = 'normal'
        if cpu_shares is not None:
            pulumi.set(__self__, "cpu_shares", cpu_shares)
        if mem_max is not None:
            pulumi.set(__self__, "mem_max", mem_max)
        if mem_min is None:
            mem_min = 200
        if mem_min is not None:
            pulumi.set(__self__, "mem_min", mem_min)
        if mem_min_expandable is None:
            mem_min_expandable = 'true'
        if mem_min_expandable is not None:
            pulumi.set(__self__, "mem_min_expandable", mem_min_expandable)
        if mem_shares is None:
            mem_shares = 'normal'
        if mem_shares is not None:
            pulumi.set(__self__, "mem_shares", mem_shares)
        if name is not None:
            pulumi.set(__self__, "name", name)

    @property
    @pulumi.getter(name="cpuMax")
    def cpu_max(self) -> Optional[pulumi.Input[int]]:
        """
        CPU maximum (in MHz).
        """
        return pulumi.get(self, "cpu_max")

    @cpu_max.setter
    def cpu_max(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "cpu_max", value)

    @property
    @pulumi.getter(name="cpuMin")
    def cpu_min(self) -> Optional[pulumi.Input[int]]:
        """
        CPU minimum (in MHz).
        """
        return pulumi.get(self, "cpu_min")

    @cpu_min.setter
    def cpu_min(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "cpu_min", value)

    @property
    @pulumi.getter(name="cpuMinExpandable")
    def cpu_min_expandable(self) -> Optional[pulumi.Input[str]]:
        """
        Can pool borrow CPU resources from parent?
        """
        return pulumi.get(self, "cpu_min_expandable")

    @cpu_min_expandable.setter
    def cpu_min_expandable(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "cpu_min_expandable", value)

    @property
    @pulumi.getter(name="cpuShares")
    def cpu_shares(self) -> Optional[pulumi.Input[str]]:
        """
        CPU shares (low/normal/high/<custom>).
        """
        return pulumi.get(self, "cpu_shares")

    @cpu_shares.setter
    def cpu_shares(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "cpu_shares", value)

    @property
    @pulumi.getter(name="memMax")
    def mem_max(self) -> Optional[pulumi.Input[int]]:
        """
        Memory maximum (in MB).
        """
        return pulumi.get(self, "mem_max")

    @mem_max.setter
    def mem_max(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "mem_max", value)

    @property
    @pulumi.getter(name="memMin")
    def mem_min(self) -> Optional[pulumi.Input[int]]:
        """
        Memory minimum (in MB).
        """
        return pulumi.get(self, "mem_min")

    @mem_min.setter
    def mem_min(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "mem_min", value)

    @property
    @pulumi.getter(name="memMinExpandable")
    def mem_min_expandable(self) -> Optional[pulumi.Input[str]]:
        """
        Can pool borrow memory resources from parent?
        """
        return pulumi.get(self, "mem_min_expandable")

    @mem_min_expandable.setter
    def mem_min_expandable(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "mem_min_expandable", value)

    @property
    @pulumi.getter(name="memShares")
    def mem_shares(self) -> Optional[pulumi.Input[str]]:
        """
        Memory shares (low/normal/high/<custom>).
        """
        return pulumi.get(self, "mem_shares")

    @mem_shares.setter
    def mem_shares(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "mem_shares", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[str]]:
        """
        Resource Pool Name
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "name", value)


class ResourcePool(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 cpu_max: Optional[pulumi.Input[int]] = None,
                 cpu_min: Optional[pulumi.Input[int]] = None,
                 cpu_min_expandable: Optional[pulumi.Input[str]] = None,
                 cpu_shares: Optional[pulumi.Input[str]] = None,
                 mem_max: Optional[pulumi.Input[int]] = None,
                 mem_min: Optional[pulumi.Input[int]] = None,
                 mem_min_expandable: Optional[pulumi.Input[str]] = None,
                 mem_shares: Optional[pulumi.Input[str]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Create a ResourcePool resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[int] cpu_max: CPU maximum (in MHz).
        :param pulumi.Input[int] cpu_min: CPU minimum (in MHz).
        :param pulumi.Input[str] cpu_min_expandable: Can pool borrow CPU resources from parent?
        :param pulumi.Input[str] cpu_shares: CPU shares (low/normal/high/<custom>).
        :param pulumi.Input[int] mem_max: Memory maximum (in MB).
        :param pulumi.Input[int] mem_min: Memory minimum (in MB).
        :param pulumi.Input[str] mem_min_expandable: Can pool borrow memory resources from parent?
        :param pulumi.Input[str] mem_shares: Memory shares (low/normal/high/<custom>).
        :param pulumi.Input[str] name: Resource Pool Name
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: Optional[ResourcePoolArgs] = None,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Create a ResourcePool resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param ResourcePoolArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(ResourcePoolArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 cpu_max: Optional[pulumi.Input[int]] = None,
                 cpu_min: Optional[pulumi.Input[int]] = None,
                 cpu_min_expandable: Optional[pulumi.Input[str]] = None,
                 cpu_shares: Optional[pulumi.Input[str]] = None,
                 mem_max: Optional[pulumi.Input[int]] = None,
                 mem_min: Optional[pulumi.Input[int]] = None,
                 mem_min_expandable: Optional[pulumi.Input[str]] = None,
                 mem_shares: Optional[pulumi.Input[str]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = ResourcePoolArgs.__new__(ResourcePoolArgs)

            __props__.__dict__["cpu_max"] = cpu_max
            if cpu_min is None:
                cpu_min = 100
            __props__.__dict__["cpu_min"] = cpu_min
            if cpu_min_expandable is None:
                cpu_min_expandable = 'true'
            __props__.__dict__["cpu_min_expandable"] = cpu_min_expandable
            if cpu_shares is None:
                cpu_shares = 'normal'
            __props__.__dict__["cpu_shares"] = cpu_shares
            __props__.__dict__["mem_max"] = mem_max
            if mem_min is None:
                mem_min = 200
            __props__.__dict__["mem_min"] = mem_min
            if mem_min_expandable is None:
                mem_min_expandable = 'true'
            __props__.__dict__["mem_min_expandable"] = mem_min_expandable
            if mem_shares is None:
                mem_shares = 'normal'
            __props__.__dict__["mem_shares"] = mem_shares
            __props__.__dict__["name"] = name
        super(ResourcePool, __self__).__init__(
            'esxi-native:index:ResourcePool',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'ResourcePool':
        """
        Get an existing ResourcePool resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = ResourcePoolArgs.__new__(ResourcePoolArgs)

        __props__.__dict__["cpu_max"] = None
        __props__.__dict__["cpu_min"] = None
        __props__.__dict__["cpu_min_expandable"] = None
        __props__.__dict__["cpu_shares"] = None
        __props__.__dict__["mem_max"] = None
        __props__.__dict__["mem_min"] = None
        __props__.__dict__["mem_min_expandable"] = None
        __props__.__dict__["mem_shares"] = None
        __props__.__dict__["name"] = None
        return ResourcePool(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter(name="cpuMax")
    def cpu_max(self) -> pulumi.Output[Optional[int]]:
        """
        CPU maximum (in MHz).
        """
        return pulumi.get(self, "cpu_max")

    @property
    @pulumi.getter(name="cpuMin")
    def cpu_min(self) -> pulumi.Output[Optional[int]]:
        """
        CPU minimum (in MHz).
        """
        return pulumi.get(self, "cpu_min")

    @property
    @pulumi.getter(name="cpuMinExpandable")
    def cpu_min_expandable(self) -> pulumi.Output[Optional[str]]:
        """
        Can pool borrow CPU resources from parent?
        """
        return pulumi.get(self, "cpu_min_expandable")

    @property
    @pulumi.getter(name="cpuShares")
    def cpu_shares(self) -> pulumi.Output[Optional[str]]:
        """
        CPU shares (low/normal/high/<custom>).
        """
        return pulumi.get(self, "cpu_shares")

    @property
    @pulumi.getter(name="memMax")
    def mem_max(self) -> pulumi.Output[Optional[int]]:
        """
        Memory maximum (in MB).
        """
        return pulumi.get(self, "mem_max")

    @property
    @pulumi.getter(name="memMin")
    def mem_min(self) -> pulumi.Output[Optional[int]]:
        """
        Memory minimum (in MB).
        """
        return pulumi.get(self, "mem_min")

    @property
    @pulumi.getter(name="memMinExpandable")
    def mem_min_expandable(self) -> pulumi.Output[Optional[str]]:
        """
        Can pool borrow memory resources from parent?
        """
        return pulumi.get(self, "mem_min_expandable")

    @property
    @pulumi.getter(name="memShares")
    def mem_shares(self) -> pulumi.Output[Optional[str]]:
        """
        Memory shares (low/normal/high/<custom>).
        """
        return pulumi.get(self, "mem_shares")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[str]:
        """
        Resource Pool Name
        """
        return pulumi.get(self, "name")

