"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Rect = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_rect_1 = require("../utils/make-rect");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element containing a rectangle.
 * @param {Number} width The width of the Rectangle
 * @param {Number} height The height of the Rectangle
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @param {null|Number} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/rect)
 */
const Rect = ({ width, edgeRoundness, height, cornerRadius, ...props }) => {
    return (jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_rect_1.makeRect)({ height, width, edgeRoundness, cornerRadius }), ...props }));
};
exports.Rect = Rect;
