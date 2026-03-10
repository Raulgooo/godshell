"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.warpPath = void 0;
const get_bounding_box_1 = require("../get-bounding-box");
const parse_path_1 = require("../parse-path");
const reduce_instructions_1 = require("../reduce-instructions");
const serialize_instructions_1 = require("../serialize-instructions");
const warp_helpers_1 = require("./warp-helpers");
const getDefaultInterpolationThreshold = (instructions) => {
    const boundingBox = (0, get_bounding_box_1.getBoundingBoxFromInstructions)(instructions);
    const longer = Math.max(boundingBox.y2 - boundingBox.y1, boundingBox.x2 - boundingBox.x1);
    return longer * 0.01;
};
/*
 * @description Allows you to remap the coordinates of an SVG using a function in order to create a warp effect.
 * @see [Documentation](https://www.remotion.dev/docs/paths/warp-path)
 */
const warpPath = (path, transformer, options) => {
    var _a;
    const reduced = (0, reduce_instructions_1.reduceInstructions)((0, parse_path_1.parsePath)(path));
    const withZFix = (0, warp_helpers_1.fixZInstruction)(reduced);
    const interpolated = (0, warp_helpers_1.svgPathInterpolate)(withZFix, (_a = options === null || options === void 0 ? void 0 : options.interpolationThreshold) !== null && _a !== void 0 ? _a : getDefaultInterpolationThreshold(withZFix));
    return (0, serialize_instructions_1.serializeInstructions)((0, warp_helpers_1.warpTransform)(interpolated, transformer));
};
exports.warpPath = warpPath;
