"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.setupOutputFileSystem = setupOutputFileSystem;
const memfs_1 = __importDefault(require("memfs"));
function setupOutputFileSystem(context) {
    const outputFileSystem = memfs_1.default.createFsFromVolume(new memfs_1.default.Volume());
    // @ts-expect-error output file sytem
    context.compiler.outputFileSystem = outputFileSystem;
    context.outputFileSystem = outputFileSystem;
}
