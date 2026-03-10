export type TransitionState = {
    entering: number;
    exiting: number;
    isInTransitionSeries: boolean;
};
export declare const useTransitionProgress: () => TransitionState;
