# coding=utf-8
# *** WARNING: this file was generated by pulumigen. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities

__all__ = ['PortGroupArgs', 'PortGroup']

@pulumi.input_type
class PortGroupArgs:
    def __init__(__self__, *,
                 v_switch: pulumi.Input[str],
                 forged_transmits: Optional[pulumi.Input[bool]] = None,
                 mac_changes: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 promiscuous_mode: Optional[pulumi.Input[bool]] = None,
                 vlan: Optional[pulumi.Input[int]] = None):
        """
        The set of arguments for constructing a PortGroup resource.
        :param pulumi.Input[str] v_switch: Virtual Switch Name.
        :param pulumi.Input[bool] forged_transmits: Forged transmits (true=Accept/false=Reject).
        :param pulumi.Input[bool] mac_changes: MAC address changes (true=Accept/false=Reject).
        :param pulumi.Input[str] name: Virtual Switch name.
        :param pulumi.Input[bool] promiscuous_mode: Promiscuous mode (true=Accept/false=Reject).
        :param pulumi.Input[int] vlan: Port Group vlan id
        """
        pulumi.set(__self__, "v_switch", v_switch)
        if forged_transmits is not None:
            pulumi.set(__self__, "forged_transmits", forged_transmits)
        if mac_changes is not None:
            pulumi.set(__self__, "mac_changes", mac_changes)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if promiscuous_mode is not None:
            pulumi.set(__self__, "promiscuous_mode", promiscuous_mode)
        if vlan is not None:
            pulumi.set(__self__, "vlan", vlan)

    @property
    @pulumi.getter(name="vSwitch")
    def v_switch(self) -> pulumi.Input[str]:
        """
        Virtual Switch Name.
        """
        return pulumi.get(self, "v_switch")

    @v_switch.setter
    def v_switch(self, value: pulumi.Input[str]):
        pulumi.set(self, "v_switch", value)

    @property
    @pulumi.getter(name="forgedTransmits")
    def forged_transmits(self) -> Optional[pulumi.Input[bool]]:
        """
        Forged transmits (true=Accept/false=Reject).
        """
        return pulumi.get(self, "forged_transmits")

    @forged_transmits.setter
    def forged_transmits(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "forged_transmits", value)

    @property
    @pulumi.getter(name="macChanges")
    def mac_changes(self) -> Optional[pulumi.Input[bool]]:
        """
        MAC address changes (true=Accept/false=Reject).
        """
        return pulumi.get(self, "mac_changes")

    @mac_changes.setter
    def mac_changes(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "mac_changes", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[str]]:
        """
        Virtual Switch name.
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter(name="promiscuousMode")
    def promiscuous_mode(self) -> Optional[pulumi.Input[bool]]:
        """
        Promiscuous mode (true=Accept/false=Reject).
        """
        return pulumi.get(self, "promiscuous_mode")

    @promiscuous_mode.setter
    def promiscuous_mode(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "promiscuous_mode", value)

    @property
    @pulumi.getter
    def vlan(self) -> Optional[pulumi.Input[int]]:
        """
        Port Group vlan id
        """
        return pulumi.get(self, "vlan")

    @vlan.setter
    def vlan(self, value: Optional[pulumi.Input[int]]):
        pulumi.set(self, "vlan", value)


class PortGroup(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 forged_transmits: Optional[pulumi.Input[bool]] = None,
                 mac_changes: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 promiscuous_mode: Optional[pulumi.Input[bool]] = None,
                 v_switch: Optional[pulumi.Input[str]] = None,
                 vlan: Optional[pulumi.Input[int]] = None,
                 __props__=None):
        """
        Create a PortGroup resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[bool] forged_transmits: Forged transmits (true=Accept/false=Reject).
        :param pulumi.Input[bool] mac_changes: MAC address changes (true=Accept/false=Reject).
        :param pulumi.Input[str] name: Virtual Switch name.
        :param pulumi.Input[bool] promiscuous_mode: Promiscuous mode (true=Accept/false=Reject).
        :param pulumi.Input[str] v_switch: Virtual Switch Name.
        :param pulumi.Input[int] vlan: Port Group vlan id
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: PortGroupArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Create a PortGroup resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param PortGroupArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(PortGroupArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 forged_transmits: Optional[pulumi.Input[bool]] = None,
                 mac_changes: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 promiscuous_mode: Optional[pulumi.Input[bool]] = None,
                 v_switch: Optional[pulumi.Input[str]] = None,
                 vlan: Optional[pulumi.Input[int]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = PortGroupArgs.__new__(PortGroupArgs)

            __props__.__dict__["forged_transmits"] = forged_transmits
            __props__.__dict__["mac_changes"] = mac_changes
            __props__.__dict__["name"] = name
            __props__.__dict__["promiscuous_mode"] = promiscuous_mode
            if v_switch is None and not opts.urn:
                raise TypeError("Missing required property 'v_switch'")
            __props__.__dict__["v_switch"] = v_switch
            __props__.__dict__["vlan"] = vlan
        super(PortGroup, __self__).__init__(
            'esxi-native:index:PortGroup',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None) -> 'PortGroup':
        """
        Get an existing PortGroup resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = PortGroupArgs.__new__(PortGroupArgs)

        __props__.__dict__["forged_transmits"] = None
        __props__.__dict__["mac_changes"] = None
        __props__.__dict__["name"] = None
        __props__.__dict__["promiscuous_mode"] = None
        __props__.__dict__["v_switch"] = None
        __props__.__dict__["vlan"] = None
        return PortGroup(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter(name="forgedTransmits")
    def forged_transmits(self) -> pulumi.Output[Optional[bool]]:
        """
        Forged transmits (true=Accept/false=Reject).
        """
        return pulumi.get(self, "forged_transmits")

    @property
    @pulumi.getter(name="macChanges")
    def mac_changes(self) -> pulumi.Output[Optional[bool]]:
        """
        MAC address changes (true=Accept/false=Reject).
        """
        return pulumi.get(self, "mac_changes")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[str]:
        """
        Port Group name.
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter(name="promiscuousMode")
    def promiscuous_mode(self) -> pulumi.Output[Optional[bool]]:
        """
        Promiscuous mode (true=Accept/false=Reject).
        """
        return pulumi.get(self, "promiscuous_mode")

    @property
    @pulumi.getter(name="vSwitch")
    def v_switch(self) -> pulumi.Output[str]:
        """
        Virtual Switch Name.
        """
        return pulumi.get(self, "v_switch")

    @property
    @pulumi.getter
    def vlan(self) -> pulumi.Output[int]:
        """
        Port Group vlan id
        """
        return pulumi.get(self, "vlan")

