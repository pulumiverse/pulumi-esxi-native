import axios from 'axios';
import * as fs from 'fs';

const controlPlaneIp = '192.168.20.20'
const talosConfigFileUrl = 'https://raw.githubusercontent.com/siderolabs/talos/master/website/content/v1.3/talos-guides/install/virtualized-platforms/vmware/cp.patch.yaml';
const talosConfigFilePath = './talosconfig';
const talosCPPatchFilePath = './cp.patch.yaml';
const talosCPFilePath = './controlplane.yaml';
const talosWorkerFilePath = './worker.yaml';
const talosctlCommand = '/usr/local/bin/talosctl';

async function downloadTalosConfigFile() {
    if (fs.existsSync(talosCPPatchFilePath)) {
        console.log('Talos config file already exists locally.');
        return;
    }

    try {
        const response = await axios.get(talosConfigFileUrl);
        fs.writeFileSync(talosCPPatchFilePath, response.data);
        console.log('Talos config file downloaded successfully.');
    } catch (error) {
        throw new Error(`Error downloading Talos config file: ${error.message}`);
    }
}

function replaceVipInConfigFile() {
    try {
        let configFileContent = fs.readFileSync(talosCPPatchFilePath, 'utf-8');
        configFileContent = configFileContent.replace(/<VIP>/g, controlPlaneIp);
        fs.writeFileSync(talosCPPatchFilePath, configFileContent);
        console.log('VIP replaced in the Talos config file.');
    } catch (error) {
        throw new Error(`Error replacing VIP in the Talos config file: ${error.message}`);
    }
}

function checkTalosctlExists() {
    try {
        fs.accessSync(`which ${talosctlCommand}`);
        console.log('talosctl exists.');
    } catch (error) {
        //throw new Error('talosctl not found. Please install talosctl and make sure it is in your system PATH.');
    }
}

function readFileAsBase64Sync(filePath: string): string {
    try {
        const data = fs.readFileSync(filePath, { encoding: 'base64' });
        return data;
    } catch (err) {
        throw new Error(`Error reading file: ${err}`);
    }
}

export async function setupTalos() {
    try {
        await downloadTalosConfigFile();
        replaceVipInConfigFile();
        checkTalosctlExists();

        if (!fs.existsSync(talosConfigFilePath) || !fs.existsSync(talosCPFilePath) || !fs.existsSync(talosWorkerFilePath)) {
            const { status, error } = require('child_process').execSync(
                `${talosctlCommand} gen config vmware-test https://${controlPlaneIp}:6443 --config-patch-control-plane @${talosCPPatchFilePath}`,
                { encoding: 'utf-8' }
            );
            if (status === 0) {
                console.log('Talos config generated successfully.');
            } else {
                throw new Error(`Error executing talosctl: ${error}`);
            }
        }
    } catch (error) {
        console.error(error);
    }
}

export function getTalosCpConfig() {
    return readFileAsBase64Sync(talosCPFilePath);
}

export function getTalosWorkerConfig() {
    return readFileAsBase64Sync(talosWorkerFilePath);
}