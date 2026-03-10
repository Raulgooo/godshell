import React from 'react';
import type { TransitionPresentation } from '../types.js';
export type WipeDirection = 'from-left' | 'from-top-left' | 'from-top' | 'from-top-right' | 'from-right' | 'from-bottom-right' | 'from-bottom' | 'from-bottom-left';
export type WipeProps = {
    direction?: WipeDirection;
    outerEnterStyle?: React.CSSProperties;
    outerExitStyle?: React.CSSProperties;
    innerEnterStyle?: React.CSSProperties;
    innerExitStyle?: React.CSSProperties;
};
export declare const wipe: (props?: WipeProps | undefined) => TransitionPresentation<WipeProps>;
