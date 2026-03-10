"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TransitionSeries = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = require("react");
const remotion_1 = require("remotion");
const no_react_1 = require("remotion/no-react");
const context_js_1 = require("./context.js");
const flatten_children_js_1 = require("./flatten-children.js");
const slide_js_1 = require("./presentations/slide.js");
const validate_js_1 = require("./validate.js");
const TransitionSeriesTransition = function (_props) {
    return null;
};
const SeriesOverlay = () => {
    return null;
};
const SeriesSequence = ({ children }) => {
    // eslint-disable-next-line react/jsx-no-useless-fragment
    return jsx_runtime_1.jsx(jsx_runtime_1.Fragment, { children: children });
};
const TransitionSeriesChildren = ({ children, }) => {
    const { fps } = (0, remotion_1.useVideoConfig)();
    const frame = (0, remotion_1.useCurrentFrame)();
    const childrenValue = (0, react_1.useMemo)(() => {
        let transitionOffsets = 0;
        let startFrame = 0;
        const flattedChildren = (0, flatten_children_js_1.flattenChildren)(children);
        // Collect overlay render info to emit after the main loop
        const overlayRenders = [];
        // Track sequence durations for overlay validation
        const sequenceDurations = [];
        let pendingOverlayValidation = false;
        const mainChildren = react_1.Children.map(flattedChildren, (child, i) => {
            var _a, _b, _c, _d, _e, _f, _g, _h, _j, _k;
            const current = child;
            if (typeof current === 'string') {
                // Don't throw if it's just some accidential whitespace
                if (current.trim() === '') {
                    return null;
                }
                throw new TypeError(`The <TransitionSeries /> component only accepts a list of <TransitionSeries.Sequence /> components as its children, but you passed a string "${current}"`);
            }
            const hasPrev = flattedChildren[i - 1];
            const nextPrev = flattedChildren[i + 1];
            const prevIsTransition = typeof hasPrev === 'string' || typeof hasPrev === 'undefined'
                ? false
                : hasPrev.type === TransitionSeriesTransition;
            const prevIsOverlay = typeof hasPrev === 'string' || typeof hasPrev === 'undefined'
                ? false
                : hasPrev.type === SeriesOverlay;
            // Handle overlay
            if (current.type === SeriesOverlay) {
                // Validate: two overlays in a row
                if (prevIsOverlay) {
                    throw new TypeError(`A <TransitionSeries.Overlay /> component must not be followed by another <TransitionSeries.Overlay /> component (nth children = ${i - 1} and ${i})`);
                }
                // Validate: overlay next to transition
                if (prevIsTransition) {
                    throw new TypeError(`A <TransitionSeries.Transition /> component must not be followed by a <TransitionSeries.Overlay /> component (nth children = ${i - 1} and ${i})`);
                }
                const nextIsTransition = typeof nextPrev === 'string' || typeof nextPrev === 'undefined'
                    ? false
                    : nextPrev.type === TransitionSeriesTransition;
                if (nextIsTransition) {
                    throw new TypeError(`A <TransitionSeries.Overlay /> component must not be followed by a <TransitionSeries.Transition /> component (nth children = ${i} and ${i + 1})`);
                }
                const overlayProps = current.props;
                (0, validate_js_1.validateDurationInFrames)(overlayProps.durationInFrames, {
                    component: `of a <TransitionSeries.Overlay /> component`,
                    allowFloats: false,
                });
                const overlayOffset = (_a = overlayProps.offset) !== null && _a !== void 0 ? _a : 0;
                if (Number.isNaN(overlayOffset)) {
                    throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must not be NaN, but got NaN.`);
                }
                if (!Number.isFinite(overlayOffset)) {
                    throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must be finite, but got ${overlayOffset}.`);
                }
                if (overlayOffset % 1 !== 0) {
                    throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must be an integer, but got ${overlayOffset}.`);
                }
                // Find the previous sequence (the cut point is at startFrame + transitionOffsets)
                const cutPoint = startFrame + transitionOffsets;
                const halfDuration = overlayProps.durationInFrames / 2;
                const overlayFrom = cutPoint - halfDuration + overlayOffset;
                if (overlayFrom < 0) {
                    throw new TypeError(`A <TransitionSeries.Overlay /> extends before frame 0. The overlay starts at frame ${overlayFrom}. Reduce the duration or adjust the offset.`);
                }
                // Validate: overlay must not exceed the previous sequence duration
                const prevSeqIdx = sequenceDurations.length - 1;
                if (prevSeqIdx >= 0) {
                    const overlayStartInPrev = halfDuration - overlayOffset;
                    if (overlayStartInPrev > sequenceDurations[prevSeqIdx]) {
                        throw new TypeError(`A <TransitionSeries.Overlay /> extends beyond the previous sequence. The overlay needs ${overlayStartInPrev} frames before the cut, but the previous sequence is only ${sequenceDurations[prevSeqIdx]} frames long.`);
                    }
                }
                // We'll validate the next sequence side after we process it
                pendingOverlayValidation = true;
                // Store overlay info for deferred rendering
                overlayRenders.push({
                    cutPoint,
                    overlayFrom,
                    durationInFrames: overlayProps.durationInFrames,
                    overlayOffset,
                    halfDuration,
                    children: overlayProps.children,
                    index: i,
                });
                return null;
            }
            if (current.type === TransitionSeriesTransition) {
                if (prevIsTransition) {
                    throw new TypeError(`A <TransitionSeries.Transition /> component must not be followed by another <TransitionSeries.Transition /> component (nth children = ${i - 1} and ${i})`);
                }
                if (prevIsOverlay) {
                    throw new TypeError(`A <TransitionSeries.Overlay /> component must not be followed by a <TransitionSeries.Transition /> component (nth children = ${i - 1} and ${i})`);
                }
                return null;
            }
            if (current.type !== SeriesSequence) {
                throw new TypeError(`The <TransitionSeries /> component only accepts a list of <TransitionSeries.Sequence />, <TransitionSeries.Transition />, and <TransitionSeries.Overlay /> components as its children, but got ${current} instead`);
            }
            const prev = typeof hasPrev === 'string' || typeof hasPrev === 'undefined'
                ? null
                : hasPrev.type === TransitionSeriesTransition
                    ? hasPrev
                    : null;
            const next = typeof nextPrev === 'string' || typeof nextPrev === 'undefined'
                ? null
                : nextPrev.type === TransitionSeriesTransition
                    ? nextPrev
                    : null;
            const castedChildAgain = current;
            const debugInfo = `index = ${i}, duration = ${castedChildAgain.props.durationInFrames}`;
            if (!(castedChildAgain === null || castedChildAgain === void 0 ? void 0 : castedChildAgain.props.children)) {
                throw new TypeError(`A <TransitionSeries.Sequence /> component (${debugInfo}) was detected to not have any children. Delete it to fix this error.`);
            }
            const durationInFramesProp = castedChildAgain.props.durationInFrames;
            const { durationInFrames, children: _children, ...passedProps } = castedChildAgain.props;
            (0, validate_js_1.validateDurationInFrames)(durationInFramesProp, {
                component: `of a <TransitionSeries.Sequence /> component`,
                allowFloats: true,
            });
            const offset = (_b = castedChildAgain.props.offset) !== null && _b !== void 0 ? _b : 0;
            if (Number.isNaN(offset)) {
                throw new TypeError(`The "offset" property of a <TransitionSeries.Sequence /> must not be NaN, but got NaN (${debugInfo}).`);
            }
            if (!Number.isFinite(offset)) {
                throw new TypeError(`The "offset" property of a <TransitionSeries.Sequence /> must be finite, but got ${offset} (${debugInfo}).`);
            }
            if (offset % 1 !== 0) {
                throw new TypeError(`The "offset" property of a <TransitionSeries.Sequence /> must be finite, but got ${offset} (${debugInfo}).`);
            }
            const currentStartFrame = startFrame + offset;
            let duration = 0;
            if (prev) {
                duration = prev.props.timing.getDurationInFrames({
                    fps,
                });
                transitionOffsets -= duration;
            }
            let actualStartFrame = currentStartFrame + transitionOffsets;
            startFrame += durationInFramesProp + offset;
            // Handle the case where the first item is a transition
            if (actualStartFrame < 0) {
                startFrame -= actualStartFrame;
                actualStartFrame = 0;
            }
            // Track sequence durations for overlay validation
            sequenceDurations.push(durationInFramesProp);
            // Validate: check if a preceding overlay extends beyond this sequence
            if (pendingOverlayValidation) {
                pendingOverlayValidation = false;
                const lastOverlay = overlayRenders[overlayRenders.length - 1];
                const framesAfterCut = lastOverlay.halfDuration + lastOverlay.overlayOffset;
                if (framesAfterCut > durationInFramesProp) {
                    throw new TypeError(`A <TransitionSeries.Overlay /> extends beyond the next sequence. The overlay needs ${framesAfterCut} frames after the cut, but the next sequence is only ${durationInFramesProp} frames long.`);
                }
            }
            const nextProgress = next
                ? next.props.timing.getProgress({
                    frame: frame -
                        actualStartFrame -
                        durationInFrames +
                        next.props.timing.getDurationInFrames({ fps }),
                    fps,
                })
                : null;
            const prevProgress = prev
                ? prev.props.timing.getProgress({
                    frame: frame - actualStartFrame,
                    fps,
                })
                : null;
            if (next &&
                durationInFramesProp < next.props.timing.getDurationInFrames({ fps })) {
                throw new Error(`The duration of a <TransitionSeries.Sequence /> must not be shorter than the duration of the next <TransitionSeries.Transition />. The transition is ${next.props.timing.getDurationInFrames({ fps })} frames long, but the sequence is only ${durationInFramesProp} frames long (${debugInfo})`);
            }
            if (prev &&
                durationInFramesProp < prev.props.timing.getDurationInFrames({ fps })) {
                throw new Error(`The duration of a <TransitionSeries.Sequence /> must not be shorter than the duration of the previous <TransitionSeries.Transition />. The transition is ${prev.props.timing.getDurationInFrames({ fps })} frames long, but the sequence is only ${durationInFramesProp} frames long (${debugInfo})`);
            }
            if (next && prev && nextProgress !== null && prevProgress !== null) {
                const nextPresentation = (_c = next.props.presentation) !== null && _c !== void 0 ? _c : (0, slide_js_1.slide)();
                const prevPresentation = (_d = prev.props.presentation) !== null && _d !== void 0 ? _d : (0, slide_js_1.slide)();
                const UppercaseNextPresentation = nextPresentation.component;
                const UppercasePrevPresentation = prevPresentation.component;
                return (jsx_runtime_1.jsx(remotion_1.Sequence
                // eslint-disable-next-line react/no-array-index-key
                , { from: actualStartFrame, durationInFrames: durationInFramesProp, ...passedProps, name: passedProps.name || '<TS.Sequence>', children: jsx_runtime_1.jsx(UppercaseNextPresentation, { passedProps: (_e = nextPresentation.props) !== null && _e !== void 0 ? _e : {}, presentationDirection: "exiting", presentationProgress: nextProgress, presentationDurationInFrames: next.props.timing.getDurationInFrames({ fps }), children: jsx_runtime_1.jsx(context_js_1.WrapInExitingProgressContext, { presentationProgress: nextProgress, children: jsx_runtime_1.jsx(UppercasePrevPresentation, { passedProps: (_f = prevPresentation.props) !== null && _f !== void 0 ? _f : {}, presentationDirection: "entering", presentationProgress: prevProgress, presentationDurationInFrames: prev.props.timing.getDurationInFrames({ fps }), children: jsx_runtime_1.jsx(context_js_1.WrapInEnteringProgressContext, { presentationProgress: prevProgress, children: child }) }) }) }) }, i));
            }
            if (prevProgress !== null && prev) {
                const prevPresentation = (_g = prev.props.presentation) !== null && _g !== void 0 ? _g : (0, slide_js_1.slide)();
                const UppercasePrevPresentation = prevPresentation.component;
                return (jsx_runtime_1.jsx(remotion_1.Sequence
                // eslint-disable-next-line react/no-array-index-key
                , { from: actualStartFrame, durationInFrames: durationInFramesProp, ...passedProps, name: passedProps.name || '<TS.Sequence>', children: jsx_runtime_1.jsx(UppercasePrevPresentation, { passedProps: (_h = prevPresentation.props) !== null && _h !== void 0 ? _h : {}, presentationDirection: "entering", presentationProgress: prevProgress, presentationDurationInFrames: prev.props.timing.getDurationInFrames({ fps }), children: jsx_runtime_1.jsx(context_js_1.WrapInEnteringProgressContext, { presentationProgress: prevProgress, children: child }) }) }, i));
            }
            if (nextProgress !== null && next) {
                const nextPresentation = (_j = next.props.presentation) !== null && _j !== void 0 ? _j : (0, slide_js_1.slide)();
                const UppercaseNextPresentation = nextPresentation.component;
                return (jsx_runtime_1.jsx(remotion_1.Sequence
                // eslint-disable-next-line react/no-array-index-key
                , { from: actualStartFrame, durationInFrames: durationInFramesProp, ...passedProps, name: passedProps.name || '<TS.Sequence>', children: jsx_runtime_1.jsx(UppercaseNextPresentation, { passedProps: (_k = nextPresentation.props) !== null && _k !== void 0 ? _k : {}, presentationDirection: "exiting", presentationProgress: nextProgress, presentationDurationInFrames: next.props.timing.getDurationInFrames({ fps }), children: jsx_runtime_1.jsx(context_js_1.WrapInExitingProgressContext, { presentationProgress: nextProgress, children: child }) }) }, i));
            }
            return (jsx_runtime_1.jsx(remotion_1.Sequence
            // eslint-disable-next-line react/no-array-index-key
            , { from: actualStartFrame, durationInFrames: durationInFramesProp, ...passedProps, name: passedProps.name || '<TS.Sequence>', children: child }, i));
        });
        // Now render overlay sequences
        const overlayElements = overlayRenders.map((overlayInfo) => {
            const info = overlayInfo;
            return (jsx_runtime_1.jsx(remotion_1.Sequence, { from: Math.round(info.overlayFrom), durationInFrames: info.durationInFrames, name: "<TS.Overlay>", layout: "absolute-fill", children: info.children }, `overlay-${info.index}`));
        });
        return [...(mainChildren || []), ...overlayElements];
    }, [children, fps, frame]);
    // eslint-disable-next-line react/jsx-no-useless-fragment
    return jsx_runtime_1.jsx(jsx_runtime_1.Fragment, { children: childrenValue });
};
/*
 * @description Manages a series of transitions and sequences for advanced animation controls in Remotion projects, handling cases with varying timings and presentations.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/transitionseries)
 */
const TransitionSeries = ({ children, name, layout: passedLayout, ...otherProps }) => {
    const displayName = name !== null && name !== void 0 ? name : '<TransitionSeries>';
    const layout = passedLayout !== null && passedLayout !== void 0 ? passedLayout : 'absolute-fill';
    if (no_react_1.NoReactInternals.ENABLE_V5_BREAKING_CHANGES &&
        layout !== 'absolute-fill') {
        throw new TypeError(`The "layout" prop of <TransitionSeries /> is not supported anymore in v5. TransitionSeries' must be absolutely positioned.`);
    }
    return (jsx_runtime_1.jsx(remotion_1.Sequence, { name: displayName, layout: layout, ...otherProps, children: jsx_runtime_1.jsx(TransitionSeriesChildren, { children: children }) }));
};
exports.TransitionSeries = TransitionSeries;
exports.TransitionSeries.Sequence = SeriesSequence;
exports.TransitionSeries.Transition = TransitionSeriesTransition;
exports.TransitionSeries.Overlay = SeriesOverlay;
remotion_1.Internals.addSequenceStackTraces(exports.TransitionSeries);
remotion_1.Internals.addSequenceStackTraces(SeriesSequence);
remotion_1.Internals.addSequenceStackTraces(SeriesOverlay);
