"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.signalRestart = exports.noOpUntilRestart = void 0;
const resolveFunctions = [];
const noOpUntilRestart = () => {
    return new Promise((resolve) => {
        resolveFunctions.push(resolve);
    });
};
exports.noOpUntilRestart = noOpUntilRestart;
const signalRestart = () => {
    resolveFunctions.forEach((f) => {
        f();
    });
    resolveFunctions.length = 0;
};
exports.signalRestart = signalRestart;
