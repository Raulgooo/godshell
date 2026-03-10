"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.resetPath = void 0;
const get_bounding_box_1 = require("./get-bounding-box");
const translate_path_1 = require("./translate-path");
/*
 * @description Translates an SVG path so that the top-left corner of the bounding box is at 0, 0.
 * @see [Documentation](https://www.remotion.dev/docs/paths/reset-path)
 */
const resetPath = (d) => {
    const box = (0, get_bounding_box_1.getBoundingBox)(d);
    return (0, translate_path_1.translatePath)(d, -box.x1, -box.y1);
};
exports.resetPath = resetPath;
