"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getSubpaths = void 0;
const construct_1 = require("./helpers/construct");
const parse_path_1 = require("./parse-path");
const serialize_instructions_1 = require("./serialize-instructions");
/*
 * @description Takes an SVG path and returns an array of subpaths, splitting the path by `M` and `m` statements.
 * @see [Documentation](https://www.remotion.dev/docs/paths/get-subpaths)
 */
const getSubpaths = (path) => {
    const parsed = (0, parse_path_1.parsePath)(path);
    const { segments } = (0, construct_1.constructFromInstructions)(parsed);
    return segments.map((seg) => {
        return (0, serialize_instructions_1.serializeInstructions)(seg);
    });
};
exports.getSubpaths = getSubpaths;
