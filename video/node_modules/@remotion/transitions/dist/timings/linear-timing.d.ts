import type { TransitionTiming } from '../types.js';
export declare const linearTiming: (options: {
    durationInFrames: number;
    easing?: ((input: number) => number) | undefined;
}) => TransitionTiming;
