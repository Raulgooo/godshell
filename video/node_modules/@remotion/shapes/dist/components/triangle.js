"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Triangle = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_triangle_1 = require("../utils/make-triangle");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element containing a triangle with same length on all sides.
 * @param {Number} length The length of one triangle side.
 * @param {string} direction The direction of the triangle
 * @param {null|Number} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/triangle)
 */
const Triangle = ({ length, direction, edgeRoundness, cornerRadius, ...props }) => {
    return (jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_triangle_1.makeTriangle)({ length, direction, edgeRoundness, cornerRadius }), ...props }));
};
exports.Triangle = Triangle;
