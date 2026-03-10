"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.none = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const NonePresentation = ({ children, presentationDirection, passedProps }) => {
    const style = (0, react_1.useMemo)(() => {
        return {
            ...(presentationDirection === 'entering'
                ? passedProps.enterStyle
                : passedProps.exitStyle),
        };
    }, [passedProps.enterStyle, passedProps.exitStyle, presentationDirection]);
    return jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children });
};
/*
 * @description A presentation that has no visual effect on its own. Instead, you control the visual effect by using the useTransitionProgress() hook.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/none)
 */
const none = (props) => {
    return {
        component: NonePresentation,
        props: props !== null && props !== void 0 ? props : {},
    };
};
exports.none = none;
