"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ensureCommand = exports.ENSURE_COMMAND = void 0;
const renderer_1 = require("@remotion/renderer");
const browser_download_bar_1 = require("../browser-download-bar");
const get_cli_options_1 = require("../get-cli-options");
const log_1 = require("../log");
const parsed_cli_1 = require("../parsed-cli");
exports.ENSURE_COMMAND = 'ensure';
const ensureCommand = async (logLevel) => {
    const indent = false;
    const { browserExecutable } = (0, get_cli_options_1.getCliOptions)({
        isStill: false,
        logLevel,
        indent,
    });
    const status = await (0, renderer_1.ensureBrowser)({
        browserExecutable,
        logLevel,
        onBrowserDownload: (0, browser_download_bar_1.defaultBrowserDownloadProgress)({
            indent,
            logLevel,
            quiet: (0, parsed_cli_1.quietFlagProvided)(),
        }),
    });
    if (status.type === 'no-browser') {
        throw new Error('should have downloaded browser');
    }
    if (status.type === 'user-defined-path') {
        log_1.Log.info({ indent, logLevel }, `Has browser at ${status.path}`);
        return;
    }
    if (status.type === 'local-puppeteer-browser') {
        log_1.Log.info({ indent, logLevel }, `Has browser at ${status.path}`);
    }
};
exports.ensureCommand = ensureCommand;
