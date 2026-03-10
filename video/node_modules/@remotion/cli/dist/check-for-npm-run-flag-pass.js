"use strict";
// If someone passes --log=verbose to npm run render
// We don't receive it.
Object.defineProperty(exports, "__esModule", { value: true });
exports.checkForNpmRunFlagPass = void 0;
const log_1 = require("./log");
const checkForNpmRunFlagPass = ({ indent, logLevel, }) => {
    if (!process.env.npm_config_log) {
        return;
    }
    log_1.Log.error({ indent, logLevel }, `The environment variable "npm_config_log" is set to "${process.env.npm_config_log}".`);
    log_1.Log.error({ indent, logLevel }, `This indicates a likely mistake:`);
    log_1.Log.error({
        indent,
        logLevel,
    }, `--log gets passed to the npm command, however npm has no "log" configuration option.`);
    log_1.Log.error({
        indent,
        logLevel,
    }, `You most likely wanted to pass --log to the Remotion CLI.`);
    log_1.Log.error({
        indent,
        logLevel,
    }, `However, arguments passed to "npm run" don't get received by the script, in this case Remotion.`);
    log_1.Log.error({
        indent,
        logLevel,
    }, `Edit the npm script and pass Remotion flags to "remotion" command instead. Example:`);
    log_1.Log.error({
        indent,
        logLevel,
    });
    log_1.Log.error({
        indent,
        logLevel,
    }, `  "render": "remotion render --log=verbose"`);
    log_1.Log.error({
        indent,
        logLevel,
    });
    process.exit(1);
};
exports.checkForNpmRunFlagPass = checkForNpmRunFlagPass;
