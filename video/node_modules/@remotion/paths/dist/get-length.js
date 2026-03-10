"use strict";
// Copied from: https://github.com/rveciana/svg-path-properties
Object.defineProperty(exports, "__esModule", { value: true });
exports.getLength = void 0;
const construct_1 = require("./helpers/construct");
/*
 * @description Gets the length of an SVG path.
 * @see [Documentation](https://www.remotion.dev/docs/paths/get-length)
 */
const getLength = (path) => {
    const constructucted = (0, construct_1.construct)(path);
    return constructucted.totalLength;
};
exports.getLength = getLength;
