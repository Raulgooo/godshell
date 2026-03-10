"use strict";
// Copied from https://stackblitz.com/edit/svg-star-generator?file=index.js
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeStar = void 0;
const paths_1 = require("@remotion/paths");
const join_points_1 = require("./join-points");
/**
 * @description Generates a star SVG path.
 * @param {Number} innerRadius The inner radius of the star.
 * @param {Number} outerRadius The outer radius of the star.
 * @param {Number} points The amount of points of the star.
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @param {null|Number} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-star)
 */
const star = ({ centerX, centerY, points, innerRadius, outerRadius, cornerRadius, edgeRoundness, }) => {
    const degreeIncrement = (Math.PI * 2) / (points * 2);
    const d = new Array(points * 2).fill(true).map((_p, i) => {
        const radius = i % 2 === 0 ? outerRadius : innerRadius;
        const angle = degreeIncrement * i - Math.PI / 2;
        const point = {
            x: centerX + radius * Math.cos(angle),
            y: centerY + radius * Math.sin(angle),
        };
        return [point.x, point.y];
    });
    return [
        ...(0, join_points_1.joinPoints)([...d, d[0]], {
            edgeRoundness,
            cornerRadius,
            roundCornerStrategy: cornerRadius > 0 ? 'bezier' : 'arc',
        }),
        { type: 'Z' },
    ];
};
const makeStar = ({ points, innerRadius, outerRadius, cornerRadius = 0, edgeRoundness = null, }) => {
    const width = outerRadius * 2;
    const height = outerRadius * 2;
    const centerX = width / 2;
    const centerY = height / 2;
    const starPathInstructions = star({
        centerX,
        centerY,
        points,
        innerRadius,
        outerRadius,
        cornerRadius,
        edgeRoundness,
    });
    const reduced = (0, paths_1.reduceInstructions)(starPathInstructions);
    const path = (0, paths_1.resetPath)((0, paths_1.serializeInstructions)(reduced));
    const boundingBox = paths_1.PathInternals.getBoundingBoxFromInstructions(reduced);
    return {
        path,
        width: boundingBox.width,
        height: boundingBox.height,
        transformOrigin: `${centerX} ${centerY}`,
        instructions: starPathInstructions,
    };
};
exports.makeStar = makeStar;
