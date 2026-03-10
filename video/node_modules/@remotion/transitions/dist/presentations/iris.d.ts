import React from 'react';
import type { TransitionPresentation } from '../types.js';
export type IrisProps = {
    width: number;
    height: number;
    outerEnterStyle?: React.CSSProperties;
    outerExitStyle?: React.CSSProperties;
    innerEnterStyle?: React.CSSProperties;
    innerExitStyle?: React.CSSProperties;
};
export declare const iris: (props: IrisProps) => TransitionPresentation<IrisProps>;
