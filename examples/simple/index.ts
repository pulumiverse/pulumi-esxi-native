import * as esxi from "@pulumi/esxi-native";

const random = new esxi.Random("my-random", { length: 24 });

export const output = random.result;
