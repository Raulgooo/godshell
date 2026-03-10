"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeEllipse = void 0;
const paths_1 = require("@remotion/paths");
/**
 * @description Generates an ellipse SVG path.
 * @param {Number} rx The radius of the ellipse on the X axis.
 * @param {Number} ry The radius of the ellipse on the Y axis.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-ellipse)
 */
const makeEllipse = ({ rx, ry }) => {
    const instructions = [
        {
            type: 'M',
            x: rx,
            y: 0,
        },
        {
            type: 'a',
            rx,
            ry,
            xAxisRotation: 0,
            largeArcFlag: true,
            sweepFlag: false,
            dx: 1,
            dy: 0,
        },
        {
            type: 'Z',
        },
    ];
    const path = (0, paths_1.serializeInstructions)(instructions);
    return {
        width: rx * 2,
        height: ry * 2,
        path,
        instructions,
        transformOrigin: `${rx} ${ry}`,
    };
};
exports.makeEllipse = makeEllipse;
