"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makePie = void 0;
const paths_1 = require("@remotion/paths");
const getCoord = ({ counterClockwise, actualProgress, rotation, radius, coord, }) => {
    const factor = counterClockwise ? -1 : 1;
    const val = Math[coord === 'x' ? 'cos' : 'sin'](factor * actualProgress * Math.PI * 2 + Math.PI * 1.5 + rotation) *
        radius +
        radius;
    const rounded = Math.round(val * 100000) / 100000;
    return rounded;
};
/**
 * @description Generates a piece of pie SVG path.
 * @param {Number} radius The radius of the circle..
 * @param {Number} progress The percentage of the circle that is filled. 0 means fully empty, 1 means fully filled.
 * @param {Boolean} closePath If set to false, no path to the middle of the circle will be drawn, leading to an open arc. Default true.
 * @param {Boolean} counterClockwise If set, the circle gets filled counterclockwise instead of clockwise. Default false.
 * @param {Number} rotation Add rotation to the path. 0 means no rotation, Math.PI * 2 means 1 full clockwise rotation
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-pie)
 */
const makePie = ({ progress, radius, closePath = true, counterClockwise = false, rotation = 0, }) => {
    const actualProgress = Math.min(Math.max(progress, 0), 1);
    const endAngleX = getCoord({
        actualProgress,
        coord: 'x',
        counterClockwise,
        radius,
        rotation,
    });
    const endAngleY = getCoord({
        actualProgress,
        coord: 'y',
        counterClockwise,
        radius,
        rotation,
    });
    const start = {
        x: getCoord({
            actualProgress: 0,
            coord: 'x',
            counterClockwise,
            radius,
            rotation,
        }),
        y: getCoord({
            actualProgress: 0,
            coord: 'y',
            counterClockwise,
            radius,
            rotation,
        }),
    };
    const end = { x: endAngleX, y: endAngleY };
    const instructions = [
        {
            type: 'M',
            ...start,
        },
        {
            type: 'A',
            rx: radius,
            ry: radius,
            xAxisRotation: 0,
            largeArcFlag: false,
            sweepFlag: !counterClockwise,
            x: actualProgress <= 0.5
                ? endAngleX
                : getCoord({
                    actualProgress: 0.5,
                    coord: 'x',
                    counterClockwise,
                    radius,
                    rotation,
                }),
            y: actualProgress <= 0.5
                ? endAngleY
                : getCoord({
                    actualProgress: 0.5,
                    coord: 'y',
                    counterClockwise,
                    radius,
                    rotation,
                }),
        },
        actualProgress > 0.5
            ? {
                type: 'A',
                rx: radius,
                ry: radius,
                xAxisRotation: 0,
                largeArcFlag: false,
                sweepFlag: !counterClockwise,
                ...end,
            }
            : null,
        actualProgress > 0 && actualProgress < 1 && closePath
            ? {
                type: 'L',
                x: radius,
                y: radius,
            }
            : null,
        closePath
            ? {
                type: 'Z',
            }
            : null,
    ].filter(Boolean);
    const path = (0, paths_1.serializeInstructions)(instructions);
    return {
        height: radius * 2,
        width: radius * 2,
        path,
        instructions,
        transformOrigin: `${radius} ${radius}`,
    };
};
exports.makePie = makePie;
