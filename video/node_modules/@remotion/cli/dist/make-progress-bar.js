"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeProgressBar = void 0;
const renderer_1 = require("@remotion/renderer");
const chalk_1 = require("./chalk");
const full = '━';
const half = '╸';
const half_right = '╺';
const totalBars = 18;
const makeProgressBar = (percentage, noColor) => {
    const color = noColor ? false : renderer_1.RenderInternals.isColorSupported();
    const barsToShow = Math.floor(percentage * totalBars);
    const extraBar = (percentage * totalBars) % barsToShow;
    const grayBars = totalBars - barsToShow;
    const showHalf = extraBar > 0.5;
    const base = full.repeat(barsToShow) + (showHalf ? half : '');
    const gray = (renderer_1.RenderInternals.isColorSupported() ? full : ' ')
        .repeat(grayBars - (showHalf ? 1 : 0))
        .split('');
    if (!showHalf && barsToShow > 0 && gray.length > 0 && color) {
        gray[0] = half_right;
    }
    if (!color) {
        return `${base}${gray.join('')}`;
    }
    return `${chalk_1.chalk.blue(base)}${chalk_1.chalk.dim(gray.join(''))}`;
};
exports.makeProgressBar = makeProgressBar;
