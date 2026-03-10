import React from 'react';
import type { TransitionPresentation } from '../types.js';
export type FlipDirection = 'from-left' | 'from-right' | 'from-top' | 'from-bottom';
export type FlipProps = {
    direction?: FlipDirection;
    perspective?: number;
    outerEnterStyle?: React.CSSProperties;
    outerExitStyle?: React.CSSProperties;
    innerEnterStyle?: React.CSSProperties;
    innerExitStyle?: React.CSSProperties;
};
export declare const flip: (props?: FlipProps | undefined) => TransitionPresentation<FlipProps>;
