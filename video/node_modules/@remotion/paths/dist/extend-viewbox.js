"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.extendViewBox = void 0;
/*
 * @description Widens an SVG viewBox in all directions by a certain scale factor.
 * @see [Documentation](https://www.remotion.dev/docs/paths/extend-viewbox)
 */
const extendViewBox = (currentViewBox, scale) => {
    const relativeScale = scale - 1;
    const splitted = currentViewBox
        .split(' ')
        .map((a) => a.trim())
        .filter((a) => a !== '')
        .map(Number);
    if (splitted.length !== 4) {
        throw new Error(`currentViewBox must be 4 valid numbers, but got "${currentViewBox}"`);
    }
    for (const part of splitted) {
        if (Number.isNaN(part)) {
            throw new Error(`currentViewBox must be 4 valid numbers, but got "${currentViewBox}"`);
        }
        if (!Number.isFinite(part)) {
            throw new Error(`currentViewBox must be 4 valid numbers, but got "${currentViewBox}"`);
        }
    }
    const [x, y, width, height] = splitted;
    return [
        x - (relativeScale * width) / 2,
        y - (relativeScale * height) / 2,
        width + relativeScale * width,
        height + relativeScale * height,
    ].join(' ');
};
exports.extendViewBox = extendViewBox;
