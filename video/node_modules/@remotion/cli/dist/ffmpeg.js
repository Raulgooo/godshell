"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ffprobeCommand = exports.ffmpegCommand = exports.dynamicLibEnv = void 0;
const renderer_1 = require("@remotion/renderer");
const client_1 = require("@remotion/renderer/client");
const node_child_process_1 = require("node:child_process");
const node_path_1 = __importDefault(require("node:path"));
const parsed_cli_1 = require("./parsed-cli");
const dynamicLibEnv = (indent, logLevel, binariesDirectory) => {
    const lib = node_path_1.default.dirname(renderer_1.RenderInternals.getExecutablePath({
        type: 'compositor',
        indent,
        logLevel,
        binariesDirectory,
    }));
    return {
        RUST_BACKTRACE: 'full',
        ...(process.platform === 'darwin'
            ? {
                DYLD_LIBRARY_PATH: lib,
            }
            : process.platform === 'win32'
                ? {
                    PATH: `${lib};${process.env.PATH}`,
                }
                : {
                    LD_LIBRARY_PATH: lib,
                }),
    };
};
exports.dynamicLibEnv = dynamicLibEnv;
const ffmpegCommand = (args, logLevel) => {
    const { value: binariesDirectory } = client_1.BrowserSafeApis.options.binariesDirectoryOption.getValue({
        commandLine: parsed_cli_1.parsedCli,
    });
    const binary = renderer_1.RenderInternals.getExecutablePath({
        type: 'ffmpeg',
        indent: false,
        logLevel,
        binariesDirectory,
    });
    renderer_1.RenderInternals.makeFileExecutableIfItIsNot(binary);
    const done = (0, node_child_process_1.spawnSync)(binary, args, {
        stdio: 'inherit',
        env: (0, exports.dynamicLibEnv)(false, logLevel, binariesDirectory),
        cwd: process.cwd(),
    });
    process.exit(done.status);
};
exports.ffmpegCommand = ffmpegCommand;
const ffprobeCommand = (args, logLevel) => {
    const { value: binariesDirectory } = client_1.BrowserSafeApis.options.binariesDirectoryOption.getValue({
        commandLine: parsed_cli_1.parsedCli,
    });
    const binary = renderer_1.RenderInternals.getExecutablePath({
        type: 'ffprobe',
        indent: false,
        logLevel,
        binariesDirectory,
    });
    renderer_1.RenderInternals.makeFileExecutableIfItIsNot(binary);
    const done = (0, node_child_process_1.spawnSync)(binary, args, {
        cwd: process.cwd(),
        stdio: 'inherit',
        env: (0, exports.dynamicLibEnv)(false, logLevel, binariesDirectory),
    });
    process.exit(done.status);
};
exports.ffprobeCommand = ffprobeCommand;
