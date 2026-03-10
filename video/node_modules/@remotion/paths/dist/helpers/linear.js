"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeLinearPosition = void 0;
const makeLinearPosition = ({ x0, x1, y0, y1, }) => {
    return {
        getTotalLength: () => {
            return Math.sqrt((x0 - x1) ** 2 + (y0 - y1) ** 2);
        },
        getPointAtLength: (pos) => {
            let fraction = pos / Math.sqrt((x0 - x1) ** 2 + (y0 - y1) ** 2);
            fraction = Number.isNaN(fraction) ? 1 : fraction;
            const newDeltaX = (x1 - x0) * fraction;
            const newDeltaY = (y1 - y0) * fraction;
            return { x: x0 + newDeltaX, y: y0 + newDeltaY };
        },
        getTangentAtLength: () => {
            const module = Math.sqrt((x1 - x0) * (x1 - x0) + (y1 - y0) * (y1 - y0));
            return { x: (x1 - x0) / module, y: (y1 - y0) / module };
        },
        type: 'linear',
    };
};
exports.makeLinearPosition = makeLinearPosition;
