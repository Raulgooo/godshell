"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeRect = void 0;
const paths_1 = require("@remotion/paths");
const join_points_1 = require("./join-points");
/**
 * @description Generates an SVG rectangle.
 * @param {Number} width The width of the rectangle.
 * @param {Number} height The height of the rectangle
 * @param {Number | null} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @param {Number} cornerRadius  * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-rect)
 */
const makeRect = ({ width, height, edgeRoundness = null, cornerRadius = 0, }) => {
    const transformOrigin = [width / 2, height / 2];
    const instructions = [
        ...(0, join_points_1.joinPoints)([
            [cornerRadius, 0],
            [width, 0],
            [width, height],
            [0, height],
            [0, 0],
        ], { edgeRoundness, cornerRadius, roundCornerStrategy: 'arc' }),
        {
            type: 'Z',
        },
    ];
    const path = (0, paths_1.serializeInstructions)(instructions);
    return {
        width,
        height,
        instructions,
        path,
        transformOrigin: transformOrigin.join(' '),
    };
};
exports.makeRect = makeRect;
