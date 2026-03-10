"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.fade = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const FadePresentation = ({ children, presentationDirection, presentationProgress, passedProps }) => {
    const isEntering = presentationDirection === 'entering';
    const style = (0, react_1.useMemo)(() => {
        return {
            opacity: isEntering
                ? presentationProgress
                : passedProps.shouldFadeOutExitingScene
                    ? 1 - presentationProgress
                    : 1,
            ...(presentationDirection === 'entering'
                ? passedProps.enterStyle
                : passedProps.exitStyle),
        };
    }, [
        isEntering,
        passedProps.enterStyle,
        passedProps.exitStyle,
        passedProps.shouldFadeOutExitingScene,
        presentationDirection,
        presentationProgress,
    ]);
    return jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children });
};
/*
 * @description Provides a simple fade transition component for sliding elements in and out.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/fade)
 */
const fade = (props) => {
    return {
        component: FadePresentation,
        props: props !== null && props !== void 0 ? props : {},
    };
};
exports.fade = fade;
