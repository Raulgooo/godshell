"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.springTiming = void 0;
const remotion_1 = require("remotion");
const springTiming = (options = {}) => {
    return {
        getDurationInFrames: ({ fps }) => {
            if (options.durationInFrames) {
                return options.durationInFrames;
            }
            return (0, remotion_1.measureSpring)({
                config: options.config,
                threshold: options.durationRestThreshold,
                fps,
            });
        },
        getProgress: ({ fps, frame }) => {
            const to = options.reverse ? 0 : 1;
            const from = options.reverse ? 1 : 0;
            return (0, remotion_1.spring)({
                fps,
                frame,
                to,
                from,
                config: options.config,
                durationInFrames: options.durationInFrames,
                durationRestThreshold: options.durationRestThreshold,
                reverse: options.reverse,
            });
        },
    };
};
exports.springTiming = springTiming;
