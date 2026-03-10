"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Circle = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_circle_1 = require("../utils/make-circle");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element drawing a circle.
 * @param {Number} radius The radius of the circle.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/circle)
 */
const Circle = ({ radius, ...props }) => {
    return jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_circle_1.makeCircle)({ radius }), ...props });
};
exports.Circle = Circle;
