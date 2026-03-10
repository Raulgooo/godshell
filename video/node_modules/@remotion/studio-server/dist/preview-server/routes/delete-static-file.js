"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.deleteStaticFileHandler = void 0;
const fs_1 = require("fs");
const path_1 = __importDefault(require("path"));
const deleteStaticFileHandler = ({ input: { relativePath }, publicDir }) => {
    const resolved = path_1.default.join(publicDir, relativePath);
    const relativeToPublicDir = path_1.default.relative(publicDir, resolved);
    if (relativeToPublicDir.startsWith('..')) {
        throw new Error(`Not allowed to write to ${relativeToPublicDir}`);
    }
    const exists = (0, fs_1.existsSync)(resolved);
    if (exists) {
        (0, fs_1.unlinkSync)(resolved);
    }
    return Promise.resolve({
        success: true,
        existed: exists,
    });
};
exports.deleteStaticFileHandler = deleteStaticFileHandler;
