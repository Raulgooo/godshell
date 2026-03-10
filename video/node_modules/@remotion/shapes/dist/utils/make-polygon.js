"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makePolygon = void 0;
const paths_1 = require("@remotion/paths");
const join_points_1 = require("./join-points");
function polygon({ points, radius, centerX, centerY, cornerRadius, edgeRoundness, }) {
    const degreeIncrement = (Math.PI * 2) / points;
    const d = new Array(points).fill(0).map((_, i) => {
        const angle = degreeIncrement * i - Math.PI / 2;
        const point = {
            x: centerX + radius * Math.cos(angle),
            y: centerY + radius * Math.sin(angle),
        };
        return [point.x, point.y];
    });
    return (0, join_points_1.joinPoints)([...d, d[0]], {
        edgeRoundness,
        cornerRadius,
        roundCornerStrategy: cornerRadius > 0 ? 'bezier' : 'arc',
    });
}
const makePolygon = ({ points, radius, cornerRadius = 0, edgeRoundness = null, }) => {
    if (points < 3) {
        throw new Error(`"points" should be minimum 3, got ${points}`);
    }
    const width = 2 * radius;
    const height = 2 * radius;
    const centerX = width / 2;
    const centerY = height / 2;
    const polygonPathInstructions = polygon({
        points,
        radius,
        centerX,
        centerY,
        cornerRadius,
        edgeRoundness,
    });
    const reduced = (0, paths_1.reduceInstructions)(polygonPathInstructions);
    const path = (0, paths_1.resetPath)((0, paths_1.serializeInstructions)(reduced));
    const boundingBox = paths_1.PathInternals.getBoundingBoxFromInstructions(reduced);
    return {
        path,
        width: boundingBox.width,
        height: boundingBox.height,
        transformOrigin: `${centerX} ${centerY}`,
        instructions: polygonPathInstructions,
    };
};
exports.makePolygon = makePolygon;
