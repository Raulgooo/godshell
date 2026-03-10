"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeHyperlink = void 0;
const is_supported_1 = require("./is-supported");
const OSC = '\u001B]';
const SEP = ';';
const BEL = '\u0007';
const makeHyperlink = ({ text, url, fallback, }) => {
    const supports = (0, is_supported_1.supportsHyperlink)();
    if (!supports) {
        return fallback;
    }
    const label = typeof text === 'function' ? text(supports) : text;
    return [OSC, '8', SEP, SEP, url, BEL, label, OSC, '8', SEP, SEP, BEL].join('');
};
exports.makeHyperlink = makeHyperlink;
