import type { TransitionPresentation } from '../types.js';
export type FadeProps = {
    enterStyle?: React.CSSProperties;
    exitStyle?: React.CSSProperties;
    shouldFadeOutExitingScene?: boolean;
};
export declare const fade: (props?: FadeProps | undefined) => TransitionPresentation<FadeProps>;
