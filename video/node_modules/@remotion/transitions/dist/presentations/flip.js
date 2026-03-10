"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.flip = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const Flip = ({ children, presentationDirection, presentationProgress, passedProps: { direction = 'from-left', perspective = 1000, innerEnterStyle, innerExitStyle, outerEnterStyle, outerExitStyle, }, }) => {
    const style = (0, react_1.useMemo)(() => {
        const startRotationEntering = direction === 'from-right' || direction === 'from-top' ? 180 : -180;
        const endRotationEntering = direction === 'from-right' || direction === 'from-top' ? -180 : 180;
        const rotation = presentationDirection === 'entering'
            ? (0, remotion_1.interpolate)(presentationProgress, [0, 1], [startRotationEntering, 0])
            : (0, remotion_1.interpolate)(presentationProgress, [0, 1], [0, endRotationEntering]);
        const rotateProperty = direction === 'from-top' || direction === 'from-bottom'
            ? 'rotateX'
            : 'rotateY';
        return {
            width: '100%',
            height: '100%',
            transform: `${rotateProperty}(${rotation}deg)`,
            backfaceVisibility: 'hidden',
            WebkitBackfaceVisibility: 'hidden',
            ...(presentationDirection === 'entering'
                ? innerEnterStyle
                : innerExitStyle),
        };
    }, [
        direction,
        innerEnterStyle,
        innerExitStyle,
        presentationDirection,
        presentationProgress,
    ]);
    const outer = (0, react_1.useMemo)(() => {
        return {
            perspective,
            // Make children also their backface hidden
            transformStyle: 'preserve-3d',
            ...(presentationDirection === 'entering'
                ? outerEnterStyle
                : outerExitStyle),
        };
    }, [outerEnterStyle, outerExitStyle, perspective, presentationDirection]);
    return (jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: outer, children: jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children }) }));
};
/*
 * @description A presentation where the exiting slide flips by 180 degrees, revealing the next slide on the back side.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/flip)
 */
const flip = (props) => {
    return { component: Flip, props: props !== null && props !== void 0 ? props : {} };
};
exports.flip = flip;
