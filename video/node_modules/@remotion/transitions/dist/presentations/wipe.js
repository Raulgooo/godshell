"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.wipe = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const makePathIn = (progress, direction) => {
    switch (direction) {
        case 'from-left':
            return `
M 0 0
L ${progress} 0
L ${progress} 1
L 0 1
Z`;
        case 'from-top-left':
            return `
M 0 0
L ${progress * 2} 0
L 0 ${progress * 2}
Z`;
        case 'from-top':
            return `
M 0 0
L 1 0
L 1 ${progress}
L 0 ${progress}
Z`;
        case 'from-top-right':
            return `
M 1 0
L ${1 - progress * 2} 0
L 1 ${progress * 2}
Z`;
        case 'from-right':
            return `
M 1 0
L 1 1
L ${1 - progress} 1
L ${1 - progress} 0
Z`;
        case 'from-bottom-right':
            return `
M 1 1
L ${1 - progress * 2} 1
L 1 ${1 - progress * 2}
Z`;
        case 'from-bottom':
            return `
M 0 1
L 1 1
L 1 ${1 - progress}
L 0 ${1 - progress}
Z`;
        case 'from-bottom-left':
            return `
M 0 1
L 0 ${1 - progress * 2}
L ${progress * 2} 1
Z`;
        default:
            throw new Error(`Unknown direction ${JSON.stringify(direction)}`);
    }
};
const makePathOut = (progress, direction) => {
    switch (direction) {
        case 'from-left':
            return `
M 1 1
L ${1 - progress} 1
L ${1 - progress} 0
L 1 0
Z`;
        case 'from-top-left':
            return `
M 1 1
L ${1 - 2 * progress} 1
L 1 ${1 - 2 * progress}
Z`;
        case 'from-top':
            return `
M 1 1
L 0 1
L 0 ${1 - progress}
L 1 ${1 - progress}
Z`;
        case 'from-top-right':
            return `
M 0 1
L ${progress * 2} 1
L 0 ${1 - progress * 2}
Z`;
        case 'from-right':
            return `
M 0 0
L ${progress} 0
L ${progress} 1
L 0 1
Z`;
        case 'from-bottom-right':
            return `
M 0 0
L ${progress * 2} 0
L 0 ${progress * 2}
Z`;
        case 'from-bottom':
            return `
M 1 0
L 0 0
L 0 ${progress}
L 1 ${progress}
Z`;
        case 'from-bottom-left':
            return `
M 1 0
L ${1 - progress * 2} 0
L 1 ${progress * 2}
Z`;
        default:
            throw new Error(`Unknown direction ${JSON.stringify(direction)}`);
    }
};
const WipePresentation = ({ children, presentationProgress, presentationDirection, passedProps: { direction = 'from-left', innerEnterStyle, innerExitStyle, outerEnterStyle, outerExitStyle, }, }) => {
    const [clipId] = (0, react_1.useState)(() => String((0, remotion_1.random)(null)));
    const progressInDirection = presentationDirection === 'entering'
        ? presentationProgress
        : 1 - presentationProgress;
    const path = presentationDirection === 'entering'
        ? makePathIn(progressInDirection, direction)
        : makePathOut(progressInDirection, direction);
    const style = (0, react_1.useMemo)(() => {
        return {
            width: '100%',
            height: '100%',
            justifyContent: 'center',
            alignItems: 'center',
            clipPath: `url(#${clipId})`,
            ...(presentationDirection === 'entering'
                ? innerEnterStyle
                : innerExitStyle),
        };
    }, [clipId, innerEnterStyle, innerExitStyle, presentationDirection]);
    const outerStyle = (0, react_1.useMemo)(() => {
        return presentationDirection === 'entering'
            ? outerEnterStyle
            : outerExitStyle;
    }, [outerEnterStyle, outerExitStyle, presentationDirection]);
    const svgStyle = (0, react_1.useMemo)(() => {
        return {
            width: '100%',
            height: '100%',
            pointerEvents: 'none',
        };
    }, []);
    return (jsx_runtime_1.jsxs(remotion_1.AbsoluteFill, { style: outerStyle, children: [
            jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children }), jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { children: jsx_runtime_1.jsx("svg", { viewBox: "0 0 1 1", style: svgStyle, children: jsx_runtime_1.jsx("defs", { children: jsx_runtime_1.jsx("clipPath", { id: clipId, clipPathUnits: "objectBoundingBox", children: jsx_runtime_1.jsx("path", { d: path, fill: "black" }) }) }) }) })
        ] }));
};
/*
 * @description A presentation where the entering slide slides over the exiting slide.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/wipe)
 */
const wipe = (props) => {
    return {
        component: WipePresentation,
        props: props !== null && props !== void 0 ? props : {},
    };
};
exports.wipe = wipe;
