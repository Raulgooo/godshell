import type { FC, PropsWithChildren } from 'react';
import type { AbsoluteFillLayout, LayoutAndStyle, SequencePropsWithoutDuration } from 'remotion';
import { NoReactInternals } from 'remotion/no-react';
import type { TransitionSeriesOverlayProps, TransitionSeriesTransitionProps } from './types.js';
declare const TransitionSeriesTransition: <PresentationProps extends Record<string, unknown>>(_props: TransitionSeriesTransitionProps<PresentationProps>) => null;
declare const SeriesOverlay: FC<TransitionSeriesOverlayProps>;
type LayoutBasedProps = true extends typeof NoReactInternals.ENABLE_V5_BREAKING_CHANGES ? AbsoluteFillLayout : LayoutAndStyle;
type SeriesSequenceProps = PropsWithChildren<{
    readonly durationInFrames: number;
    readonly offset?: number;
    readonly className?: string;
    /**
     * @deprecated For internal use only
     */
    readonly stack?: string;
} & LayoutBasedProps & Pick<SequencePropsWithoutDuration, 'name'>>;
declare const SeriesSequence: ({ children }: SeriesSequenceProps) => import("react/jsx-runtime").JSX.Element;
export declare const TransitionSeries: FC<SequencePropsWithoutDuration> & {
    Sequence: typeof SeriesSequence;
    Transition: typeof TransitionSeriesTransition;
    Overlay: typeof SeriesOverlay;
};
export {};
