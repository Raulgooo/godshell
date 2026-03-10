"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.slide = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const epsilon = 0.01;
const SlidePresentation = ({ children, presentationProgress, presentationDirection, passedProps: { direction = 'from-left', enterStyle, exitStyle }, }) => {
    const directionStyle = (0, react_1.useMemo)(() => {
        // Overlay the two slides barely to avoid a white line between them
        // Remove the correction once the presentation progress is 1
        const presentationProgressWithEpsilonCorrection = presentationProgress === 1
            ? presentationProgress * 100
            : presentationProgress * 100 - epsilon;
        if (presentationDirection === 'exiting') {
            switch (direction) {
                case 'from-left':
                    return {
                        transform: `translateX(${presentationProgressWithEpsilonCorrection}%)`,
                    };
                case 'from-right':
                    return {
                        transform: `translateX(${-presentationProgress * 100}%)`,
                    };
                case 'from-top':
                    return {
                        transform: `translateY(${presentationProgressWithEpsilonCorrection}%)`,
                    };
                case 'from-bottom':
                    return {
                        transform: `translateY(${-presentationProgress * 100}%)`,
                    };
                default:
                    throw new Error(`Invalid direction: ${direction}`);
            }
        }
        switch (direction) {
            case 'from-left':
                return {
                    transform: `translateX(${-100 + presentationProgress * 100}%)`,
                };
            case 'from-right':
                return {
                    transform: `translateX(${100 - presentationProgressWithEpsilonCorrection}%)`,
                };
            case 'from-top':
                return {
                    transform: `translateY(${-100 + presentationProgress * 100}%)`,
                };
            case 'from-bottom':
                return {
                    transform: `translateY(${100 - presentationProgressWithEpsilonCorrection}%)`,
                };
            default:
                throw new Error(`Invalid direction: ${direction}`);
        }
    }, [presentationDirection, presentationProgress, direction]);
    const style = (0, react_1.useMemo)(() => {
        return {
            width: '100%',
            height: '100%',
            justifyContent: 'center',
            alignItems: 'center',
            ...directionStyle,
            ...(presentationDirection === 'entering' ? enterStyle : exitStyle),
        };
    }, [directionStyle, enterStyle, exitStyle, presentationDirection]);
    return jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children });
};
/*
 * @description Implements a sliding transition for presentation components where the entering slide pushes the outgoing slide in a specified direction.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/slide)
 */
const slide = (props) => {
    return {
        component: SlidePresentation,
        props: props !== null && props !== void 0 ? props : {},
    };
};
exports.slide = slide;
