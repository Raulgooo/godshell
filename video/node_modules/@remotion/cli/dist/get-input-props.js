"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getInputProps = void 0;
const node_fs_1 = __importDefault(require("node:fs"));
const node_os_1 = __importDefault(require("node:os"));
const node_path_1 = __importDefault(require("node:path"));
const log_1 = require("./log");
const parsed_cli_1 = require("./parsed-cli");
const getInputProps = (onUpdate, logLevel) => {
    if (!parsed_cli_1.parsedCli.props) {
        return {};
    }
    const jsonFile = node_path_1.default.resolve(process.cwd(), parsed_cli_1.parsedCli.props);
    try {
        if (node_fs_1.default.existsSync(jsonFile)) {
            const rawJsonData = node_fs_1.default.readFileSync(jsonFile, 'utf-8');
            if (onUpdate) {
                node_fs_1.default.watchFile(jsonFile, { interval: 100 }, () => {
                    try {
                        onUpdate(JSON.parse(node_fs_1.default.readFileSync(jsonFile, 'utf-8')));
                        log_1.Log.info({ indent: false, logLevel }, `Updated input props from ${jsonFile}.`);
                    }
                    catch (_a) {
                        log_1.Log.error({ indent: false, logLevel }, `${jsonFile} contains invalid JSON. Did not apply new input props.`);
                    }
                });
            }
            return JSON.parse(rawJsonData);
        }
        return JSON.parse(parsed_cli_1.parsedCli.props);
    }
    catch (_a) {
        log_1.Log.error({ indent: false, logLevel }, 'You passed --props but it was neither valid JSON nor a file path to a valid JSON file. Provided value: ' +
            parsed_cli_1.parsedCli.props);
        log_1.Log.info({ indent: false, logLevel }, 'Got the following value:', parsed_cli_1.parsedCli.props);
        log_1.Log.error({ indent: false, logLevel }, 'Check that your input is parseable using `JSON.parse` and try again.');
        if (node_os_1.default.platform() === 'win32') {
            const logOptions = {
                indent: false,
                logLevel,
            };
            log_1.Log.warn(logOptions, 'Note: Windows handles escaping of quotes very weirdly in the command line.');
            log_1.Log.warn(logOptions, 'This might have led to you having this problem.');
            log_1.Log.warn(logOptions, 'Consider using the alternative API for --props which is to pass');
            log_1.Log.warn(logOptions, 'a path to a JSON file:');
            log_1.Log.warn(logOptions, '  --props=path/to/props.json');
        }
        process.exit(1);
    }
};
exports.getInputProps = getInputProps;
