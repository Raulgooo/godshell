"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.upgradeCommand = void 0;
const renderer_1 = require("@remotion/renderer");
const studio_server_1 = require("@remotion/studio-server");
const node_child_process_1 = require("node:child_process");
const chalk_1 = require("./chalk");
const list_of_remotion_packages_1 = require("./list-of-remotion-packages");
const log_1 = require("./log");
const upgradeCommand = async ({ remotionRoot, packageManager, version, logLevel, args, }) => {
    const { dependencies, devDependencies, optionalDependencies, peerDependencies, } = studio_server_1.StudioServerInternals.getInstalledDependencies(remotionRoot);
    let targetVersion;
    if (version) {
        targetVersion = version;
        log_1.Log.info({ indent: false, logLevel }, 'Upgrading to specified version: ' + version);
    }
    else {
        targetVersion = await studio_server_1.StudioServerInternals.getLatestRemotionVersion();
        log_1.Log.info({ indent: false, logLevel }, 'Newest Remotion version is', targetVersion);
    }
    const manager = studio_server_1.StudioServerInternals.getPackageManager(remotionRoot, packageManager, 0);
    if (manager === 'unknown') {
        throw new Error(`No lockfile was found in your project (one of ${studio_server_1.StudioServerInternals.lockFilePaths
            .map((p) => p.path)
            .join(', ')}). Install dependencies using your favorite manager!`);
    }
    const toUpgrade = list_of_remotion_packages_1.listOfRemotionPackages.filter((u) => dependencies.includes(u) ||
        devDependencies.includes(u) ||
        optionalDependencies.includes(u) ||
        peerDependencies.includes(u));
    const command = studio_server_1.StudioServerInternals.getInstallCommand({
        manager: manager.manager,
        packages: toUpgrade,
        version: targetVersion,
        additionalArgs: args,
    });
    log_1.Log.info({ indent: false, logLevel }, chalk_1.chalk.gray(`$ ${manager.manager} ${command.join(' ')}`));
    const task = (0, node_child_process_1.spawn)(manager.manager, command, {
        env: {
            ...process.env,
            ADBLOCK: '1',
            DISABLE_OPENCOLLECTIVE: '1',
        },
        stdio: renderer_1.RenderInternals.isEqualOrBelowLogLevel(logLevel, 'info')
            ? 'inherit'
            : 'ignore',
    });
    await new Promise((resolve) => {
        task.on('close', (code) => {
            if (code === 0) {
                resolve();
            }
            else if (renderer_1.RenderInternals.isEqualOrBelowLogLevel(logLevel, 'info')) {
                throw new Error('Failed to upgrade Remotion, see logs above');
            }
            else {
                throw new Error('Failed to upgrade Remotion, run with --log=info info to see logs');
            }
        });
    });
    log_1.Log.info({ indent: false, logLevel }, '⏫ Remotion has been upgraded!');
    log_1.Log.info({ indent: false, logLevel }, 'https://remotion.dev/changelog');
};
exports.upgradeCommand = upgradeCommand;
