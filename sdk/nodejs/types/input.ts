// *** WARNING: this file was generated by pulumigen. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import { input as inputs, output as outputs, enums } from "../types";

export interface KeyValuePairArgs {
    key?: pulumi.Input<string>;
    value?: pulumi.Input<string>;
}

export interface NetworkInterfaceArgs {
    macAddress?: pulumi.Input<string>;
    nicType?: pulumi.Input<string>;
    virtualNetwork?: pulumi.Input<string>;
}

export interface UplinkArgs {
    /**
     * Uplink name.
     */
    name: pulumi.Input<string>;
}

export interface VMVirtualDiskArgs {
    /**
     * SCSI_Ctrl:SCSI_id.    Range  '0:1' to '0:15'.   SCSI_id 7 is not allowed.
     */
    slot?: pulumi.Input<string>;
    virtualDiskId?: pulumi.Input<string>;
}
