"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getInstalledInstallablePackages = void 0;
const studio_shared_1 = require("@remotion/studio-shared");
const get_installed_dependencies_1 = require("./get-installed-dependencies");
const getInstalledInstallablePackages = (remotionRoot) => {
    const { dependencies, devDependencies, optionalDependencies } = (0, get_installed_dependencies_1.getInstalledDependencies)(remotionRoot);
    const installablePackages = [
        ...dependencies,
        ...devDependencies,
        ...optionalDependencies,
    ];
    return Object.entries(studio_shared_1.installableMap)
        .filter(([, _installable]) => _installable)
        .map(([pkg]) => (pkg === 'core' ? 'remotion' : `@remotion/${pkg}`))
        .filter((pkg) => installablePackages.includes(pkg));
};
exports.getInstalledInstallablePackages = getInstalledInstallablePackages;
