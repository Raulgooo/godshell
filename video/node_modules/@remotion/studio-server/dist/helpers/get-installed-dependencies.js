"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getInstalledDependencies = void 0;
const node_fs_1 = __importDefault(require("node:fs"));
const node_path_1 = __importDefault(require("node:path"));
const getInstalledDependencies = (remotionRoot) => {
    var _a, _b, _c, _d;
    const packageJsonFilePath = node_path_1.default.join(remotionRoot, 'package.json');
    const packageJson = JSON.parse(node_fs_1.default.readFileSync(packageJsonFilePath, 'utf-8'));
    const dependencies = Object.keys((_a = packageJson.dependencies) !== null && _a !== void 0 ? _a : {});
    const devDependencies = Object.keys((_b = packageJson.devDependencies) !== null && _b !== void 0 ? _b : {});
    const optionalDependencies = Object.keys((_c = packageJson.optionalDependencies) !== null && _c !== void 0 ? _c : {});
    const peerDependencies = Object.keys((_d = packageJson.peerDependencies) !== null && _d !== void 0 ? _d : {});
    return {
        dependencies,
        devDependencies,
        optionalDependencies,
        peerDependencies,
    };
};
exports.getInstalledDependencies = getInstalledDependencies;
