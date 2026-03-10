"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Ellipse = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_ellipse_1 = require("../utils/make-ellipse");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element drawing an ellipse.
 * @param {Number} rx The radius of the ellipse on the X axis.
 * @param {Number} ry The radius of the ellipse on the Y axis.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/ellipse)
 */
const Ellipse = ({ rx, ry, ...props }) => {
    return jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_ellipse_1.makeEllipse)({ rx, ry }), ...props });
};
exports.Ellipse = Ellipse;
