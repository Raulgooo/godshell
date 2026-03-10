import React from 'react';
import type { TransitionPresentation } from '../types.js';
export type SlideDirection = 'from-left' | 'from-top' | 'from-right' | 'from-bottom';
export type SlideProps = {
    direction?: SlideDirection;
    exitStyle?: React.CSSProperties;
    enterStyle?: React.CSSProperties;
};
export declare const slide: (props?: SlideProps | undefined) => TransitionPresentation<SlideProps>;
