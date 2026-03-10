"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.translatePath = exports.translateSegments = void 0;
const parse_path_1 = require("./parse-path");
const serialize_instructions_1 = require("./serialize-instructions");
const translateSegments = (segments, x, y) => {
    return segments.map((segment) => {
        // Shift coords only for commands with absolute values
        if (segment.type === 'a' ||
            segment.type === 'c' ||
            segment.type === 'v' ||
            segment.type === 's' ||
            segment.type === 'h' ||
            segment.type === 'l' ||
            segment.type === 'm' ||
            segment.type === 'q' ||
            segment.type === 't') {
            return segment;
        }
        // V is the only command, with shifted coords parity
        if (segment.type === 'V') {
            return {
                type: 'V',
                y: segment.y + y,
            };
        }
        if (segment.type === 'H') {
            return {
                type: 'H',
                x: segment.x + x,
            };
        }
        // ARC is: ['A', rx, ry, x-axis-rotation, large-arc-flag, sweep-flag, x, y]
        // touch x, y only
        if (segment.type === 'A') {
            return {
                type: 'A',
                rx: segment.rx,
                ry: segment.ry,
                largeArcFlag: segment.largeArcFlag,
                sweepFlag: segment.sweepFlag,
                xAxisRotation: segment.xAxisRotation,
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'Z') {
            return segment;
        }
        if (segment.type === 'C') {
            return {
                type: 'C',
                cp1x: segment.cp1x + x,
                cp1y: segment.cp1y + y,
                cp2x: segment.cp2x + x,
                cp2y: segment.cp2y + y,
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'Q') {
            return {
                type: 'Q',
                cpx: segment.cpx + x,
                cpy: segment.cpy + y,
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'S') {
            return {
                type: 'S',
                cpx: segment.cpx + x,
                cpy: segment.cpy + y,
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'T') {
            return {
                type: 'T',
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'L') {
            return {
                type: 'L',
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        if (segment.type === 'M') {
            return {
                type: 'M',
                x: segment.x + x,
                y: segment.y + y,
            };
        }
        // @ts-expect-error
        throw new Error(`Unknown segment type: ${segment.type}`);
    });
};
exports.translateSegments = translateSegments;
/*
 * @description Translates the path by the given x and y coordinates.
 * @see [Documentation](https://www.remotion.dev/docs/paths/translate-path)
 */
const translatePath = (path, x, y) => {
    return (0, serialize_instructions_1.serializeInstructions)((0, exports.translateSegments)((0, parse_path_1.parsePath)(path), x, y));
};
exports.translatePath = translatePath;
