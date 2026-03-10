"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handleInstallPackage = void 0;
const renderer_1 = require("@remotion/renderer");
const node_child_process_1 = require("node:child_process");
const version_1 = require("remotion/version");
const install_command_1 = require("../../helpers/install-command");
const get_package_manager_1 = require("../get-package-manager");
const handleInstallPackage = async ({ logLevel, remotionRoot, input: { packageNames } }) => {
    for (const packageName of packageNames) {
        if (!packageName.startsWith('@remotion/')) {
            return Promise.reject(new Error(`Package ${packageName} is not allowed to be installed.`));
        }
    }
    const manager = (0, get_package_manager_1.getPackageManager)(remotionRoot, undefined, 0);
    if (manager === 'unknown') {
        throw new Error(`No lockfile was found in your project (one of ${get_package_manager_1.lockFilePaths
            .map((p) => p.path)
            .join(', ')}). Install dependencies using your favorite manager!`);
    }
    const command = (0, install_command_1.getInstallCommand)({
        manager: manager.manager,
        packages: packageNames,
        version: version_1.VERSION,
        additionalArgs: [],
    });
    renderer_1.RenderInternals.Log.info({ indent: false, logLevel }, renderer_1.RenderInternals.chalk.gray(`╭─  ${manager.manager} ${command.join(' ')}`));
    const time = Date.now();
    try {
        await new Promise((resolve, reject) => {
            const cmd = (0, node_child_process_1.spawn)(manager.manager, command, {});
            cmd.stdout.on('data', (d) => {
                const splitted = d.toString().trim().split('\n');
                splitted.forEach((line) => {
                    renderer_1.RenderInternals.Log.info({ indent: true, logLevel }, line);
                });
            });
            cmd.stdout.on('end', () => {
                resolve();
            });
            cmd.on('close', (code, signal) => {
                if (code === 0) {
                    resolve();
                }
                else {
                    reject(new Error(`Command exited with code ${code} and signal ${signal}`));
                }
            });
        });
        const timeEnd = Date.now();
        renderer_1.RenderInternals.Log.info({ indent: false, logLevel }, renderer_1.RenderInternals.chalk.gray('╰─ '), `Done in ${timeEnd - time}ms`);
        return Promise.resolve({});
    }
    catch (err) {
        const timeEnd = Date.now();
        renderer_1.RenderInternals.Log.info({ indent: false, logLevel }, renderer_1.RenderInternals.chalk.gray('╰─ '), renderer_1.RenderInternals.chalk.red(`Errored in ${timeEnd - time}ms`));
        return Promise.reject(err);
    }
};
exports.handleInstallPackage = handleInstallPackage;
