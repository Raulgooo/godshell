"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.iris = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const paths_1 = require("@remotion/paths");
const shapes_1 = require("@remotion/shapes");
const react_1 = require("react");
const remotion_1 = require("remotion");
const IrisPresentation = ({ children, presentationDirection, presentationProgress, passedProps }) => {
    // Calculate the radius needed to cover the entire viewport
    const maxRadius = Math.sqrt(passedProps.width ** 2 + passedProps.height ** 2) / 2;
    // For iris, we want to start small and grow large (entering)
    // or start large and shrink small (exiting)
    const minRadius = 0;
    const currentRadius = presentationDirection === 'entering'
        ? minRadius + (maxRadius - minRadius) * presentationProgress
        : maxRadius - (maxRadius - minRadius) * presentationProgress;
    const { path } = (0, shapes_1.makeCircle)({
        radius: currentRadius,
    });
    // Center the circle in the viewport
    const translatedPath = (0, paths_1.translatePath)(path, passedProps.width / 2 - currentRadius, passedProps.height / 2 - currentRadius);
    const [clipId] = (0, react_1.useState)(() => String((0, remotion_1.random)(null)));
    const style = (0, react_1.useMemo)(() => {
        return {
            width: '100%',
            height: '100%',
            clipPath: presentationDirection === 'exiting' ? undefined : `url(#${clipId})`,
            ...(presentationDirection === 'entering'
                ? passedProps.innerEnterStyle
                : passedProps.innerExitStyle),
        };
    }, [
        clipId,
        passedProps.innerEnterStyle,
        passedProps.innerExitStyle,
        presentationDirection,
    ]);
    const outerStyle = (0, react_1.useMemo)(() => {
        return presentationDirection === 'entering'
            ? passedProps.outerEnterStyle
            : passedProps.outerExitStyle;
    }, [
        passedProps.outerEnterStyle,
        passedProps.outerExitStyle,
        presentationDirection,
    ]);
    return (jsx_runtime_1.jsxs(remotion_1.AbsoluteFill, { style: outerStyle, children: [
            jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { style: style, children: children }), presentationDirection === 'exiting' ? null : (jsx_runtime_1.jsx(remotion_1.AbsoluteFill, { children: jsx_runtime_1.jsx("svg", { children: jsx_runtime_1.jsx("defs", { children: jsx_runtime_1.jsx("clipPath", { id: clipId, children: jsx_runtime_1.jsx("path", { d: translatedPath, fill: "black" }) }) }) }) }))] }));
};
/*
 * @description Creates an iris transition that uses a circular mask starting from the center to reveal the underlying scene.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/presentations/iris)
 */
const iris = (props) => {
    return { component: IrisPresentation, props: props !== null && props !== void 0 ? props : {} };
};
exports.iris = iris;
