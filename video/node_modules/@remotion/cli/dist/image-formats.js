"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getVideoImageFormat = void 0;
const renderer_1 = require("@remotion/renderer");
const pure_1 = require("@remotion/renderer/pure");
const config_1 = require("./config");
const parsed_cli_1 = require("./parsed-cli");
const getVideoImageFormat = ({ codec, uiImageFormat, }) => {
    if (uiImageFormat !== null) {
        return uiImageFormat;
    }
    if (typeof parsed_cli_1.parsedCli['image-format'] !== 'undefined') {
        if (!renderer_1.RenderInternals.validVideoImageFormats.includes(parsed_cli_1.parsedCli['image-format'])) {
            throw new Error(`Invalid image format: ${parsed_cli_1.parsedCli['image-format']}`);
        }
        return parsed_cli_1.parsedCli['image-format'];
    }
    const configFileOption = config_1.ConfigInternals.getUserPreferredVideoImageFormat();
    if (typeof configFileOption !== 'undefined') {
        return configFileOption;
    }
    if (pure_1.NoReactAPIs.isAudioCodec(codec)) {
        return 'none';
    }
    if (codec === 'h264' ||
        codec === 'h264-mkv' ||
        codec === 'h264-ts' ||
        codec === 'h265' ||
        codec === 'vp8' ||
        codec === 'vp9' ||
        codec === 'prores' ||
        codec === 'gif') {
        return 'jpeg';
    }
    if (codec === undefined) {
        return 'png';
    }
    throw new Error('Unrecognized codec ' + codec);
};
exports.getVideoImageFormat = getVideoImageFormat;
