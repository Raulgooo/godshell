import React from 'react';
import type { TransitionPresentation } from '../types.js';
export type ClockWipeProps = {
    width: number;
    height: number;
    outerEnterStyle?: React.CSSProperties;
    outerExitStyle?: React.CSSProperties;
    innerEnterStyle?: React.CSSProperties;
    innerExitStyle?: React.CSSProperties;
};
export declare const clockWipe: (props: ClockWipeProps) => TransitionPresentation<ClockWipeProps>;
