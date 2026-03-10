"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handleRestartStudio = void 0;
const close_and_restart_1 = require("../close-and-restart");
const handleRestartStudio = () => {
    (0, close_and_restart_1.signalRestart)();
    return Promise.resolve({});
};
exports.handleRestartStudio = handleRestartStudio;
