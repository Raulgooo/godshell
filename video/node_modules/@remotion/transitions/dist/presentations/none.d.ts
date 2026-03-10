import React from 'react';
import type { TransitionPresentation } from '../types';
export type NoneProps = {
    enterStyle?: React.CSSProperties;
    exitStyle?: React.CSSProperties;
};
export declare const none: (props?: NoneProps | undefined) => TransitionPresentation<NoneProps>;
