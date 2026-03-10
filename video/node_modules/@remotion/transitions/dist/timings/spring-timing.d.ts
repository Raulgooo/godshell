import type { SpringConfig } from 'remotion';
import type { TransitionTiming } from '../types.js';
export declare const springTiming: (options?: {
    config?: Partial<SpringConfig> | undefined;
    durationInFrames?: number | undefined;
    durationRestThreshold?: number | undefined;
    reverse?: boolean | undefined;
}) => TransitionTiming;
