"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.initializeCli = void 0;
const client_1 = require("@remotion/renderer/client");
const path_1 = __importDefault(require("path"));
const get_config_file_name_1 = require("./get-config-file-name");
const make_link_1 = require("./hyperlinks/make-link");
const log_1 = require("./log");
const parse_command_line_1 = require("./parse-command-line");
const parsed_cli_1 = require("./parsed-cli");
const initializeCli = async (remotionRoot) => {
    const appliedName = await (0, get_config_file_name_1.loadConfig)(remotionRoot);
    (0, parse_command_line_1.parseCommandLine)();
    const logLevel = client_1.BrowserSafeApis.options.logLevelOption.getValue({
        commandLine: parsed_cli_1.parsedCli,
    }).value;
    // Only now Log.verbose is available
    log_1.Log.verbose({ indent: false, logLevel }, 'Remotion root directory:', (0, make_link_1.makeHyperlink)({
        fallback: remotionRoot,
        text: path_1.default.relative(process.cwd(), remotionRoot) || '.',
        url: `file://${remotionRoot}`,
    }));
    if (remotionRoot !== process.cwd()) {
        log_1.Log.warn({ indent: false, logLevel }, `Warning: The root directory of your project is ${remotionRoot}, but you are executing this command from ${process.cwd()}. The recommendation is to execute commands from the root directory.`);
    }
    if (appliedName) {
        log_1.Log.verbose({ indent: false, logLevel }, `Applied configuration from ${(0, make_link_1.makeHyperlink)({ url: `file://${appliedName}`, text: path_1.default.relative(process.cwd(), appliedName), fallback: appliedName })}.`);
    }
    else {
        log_1.Log.verbose({ indent: false, logLevel }, 'No config file loaded.');
    }
    return logLevel;
};
exports.initializeCli = initializeCli;
