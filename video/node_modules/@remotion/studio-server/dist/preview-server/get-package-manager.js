"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getPackageManager = exports.lockFilePaths = void 0;
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
exports.lockFilePaths = [
    {
        path: 'package-lock.json',
        manager: 'npm',
        installCommand: 'npm i',
        startCommand: 'npx remotion studio',
    },
    {
        path: 'yarn.lock',
        manager: 'yarn',
        installCommand: 'yarn add',
        startCommand: 'yarn remotion studio',
    },
    {
        path: 'pnpm-lock.yaml',
        manager: 'pnpm',
        installCommand: 'pnpm i',
        startCommand: 'pnpm exec remotion studio',
    },
    {
        path: 'bun.lock',
        manager: 'bun',
        installCommand: 'bun i',
        startCommand: 'bunx remotion studio',
    },
    {
        path: 'bun.lockb',
        manager: 'bun',
        installCommand: 'bun i',
        startCommand: 'bunx remotion studio',
    },
];
const getPackageManager = (remotionRoot, packageManager, dirUp) => {
    if (packageManager) {
        const manager = exports.lockFilePaths.find((p) => p.manager === packageManager);
        if (!manager) {
            throw new Error(`The package manager ${packageManager} is not supported. Supported package managers are ${exports.lockFilePaths
                .map((p) => p.manager)
                .join(', ')}`);
        }
        return manager;
    }
    const existingPkgManagers = exports.lockFilePaths.filter((p) => node_fs_1.default.existsSync(node_path_1.default.join(remotionRoot, ...new Array(dirUp).fill('..'), p.path)));
    if (existingPkgManagers.length === 0 && dirUp >= 2) {
        return 'unknown';
    }
    if (existingPkgManagers.length === 0) {
        return (0, exports.getPackageManager)(remotionRoot, packageManager, dirUp + 1);
    }
    if (existingPkgManagers.length > 1) {
        const error = [
            `Found multiple lockfiles:`,
            ...existingPkgManagers.map((m) => {
                return `- ${m.path}`;
            }),
            '',
            'This can lead to bugs, delete all but one of these files and run this command again.',
        ].join('\n');
        throw new Error(error);
    }
    return existingPkgManagers[0];
};
exports.getPackageManager = getPackageManager;
