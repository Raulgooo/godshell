"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Star = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_star_1 = require("../utils/make-star");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element containing a star.
 * @param {Number} innerRadius The inner radius of the star.
 * @param {Number} outerRadius The outer radius of the star.
 * @param {Number} points The amount of points of the star.
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @param {null|Number} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/star)
 */
const Star = ({ innerRadius, outerRadius, points, cornerRadius, edgeRoundness, ...props }) => {
    return (jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_star_1.makeStar)({
            innerRadius,
            outerRadius,
            points,
            cornerRadius,
            edgeRoundness,
        }), ...props }));
};
exports.Star = Star;
