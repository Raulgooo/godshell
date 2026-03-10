"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeCircle = void 0;
const paths_1 = require("@remotion/paths");
/**
 * @description Generates a circle SVG path.
 * @param {Number} radius The radius of the circle.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-circle)
 */
const makeCircle = ({ radius }) => {
    const instructions = [
        {
            type: 'M',
            x: radius,
            y: 0,
        },
        {
            type: 'a',
            rx: radius,
            ry: radius,
            xAxisRotation: 0,
            largeArcFlag: true,
            sweepFlag: true,
            dx: 0,
            dy: radius * 2,
        },
        {
            type: 'a',
            rx: radius,
            ry: radius,
            xAxisRotation: 0,
            largeArcFlag: true,
            sweepFlag: true,
            dx: 0,
            dy: -radius * 2,
        },
        {
            type: 'Z',
        },
    ];
    const path = (0, paths_1.serializeInstructions)(instructions);
    return {
        height: radius * 2,
        width: radius * 2,
        path,
        instructions,
        transformOrigin: `${radius} ${radius}`,
    };
};
exports.makeCircle = makeCircle;
