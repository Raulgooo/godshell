"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Heart = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_heart_1 = require("../utils/make-heart");
const render_svg_1 = require("./render-svg");
/**
 * @description Renders an SVG element containing a heart.
 * @param {Number} size The size of the heart.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/heart)
 */
const Heart = ({ aspectRatio, height, bottomRoundnessAdjustment = 0, depthAdjustment = 0, ...props }) => {
    return (jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_heart_1.makeHeart)({
            aspectRatio,
            height,
            bottomRoundnessAdjustment,
            depthAdjustment,
        }), ...props }));
};
exports.Heart = Heart;
