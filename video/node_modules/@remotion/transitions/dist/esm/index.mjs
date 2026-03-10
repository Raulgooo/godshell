// src/presentations/slide.tsx
import { useMemo } from "react";
import { AbsoluteFill } from "remotion";
import { jsx } from "react/jsx-runtime";
var epsilon = 0.01;
var SlidePresentation = ({
  children,
  presentationProgress,
  presentationDirection,
  passedProps: { direction = "from-left", enterStyle, exitStyle }
}) => {
  const directionStyle = useMemo(() => {
    const presentationProgressWithEpsilonCorrection = presentationProgress === 1 ? presentationProgress * 100 : presentationProgress * 100 - epsilon;
    if (presentationDirection === "exiting") {
      switch (direction) {
        case "from-left":
          return {
            transform: `translateX(${presentationProgressWithEpsilonCorrection}%)`
          };
        case "from-right":
          return {
            transform: `translateX(${-presentationProgress * 100}%)`
          };
        case "from-top":
          return {
            transform: `translateY(${presentationProgressWithEpsilonCorrection}%)`
          };
        case "from-bottom":
          return {
            transform: `translateY(${-presentationProgress * 100}%)`
          };
        default:
          throw new Error(`Invalid direction: ${direction}`);
      }
    }
    switch (direction) {
      case "from-left":
        return {
          transform: `translateX(${-100 + presentationProgress * 100}%)`
        };
      case "from-right":
        return {
          transform: `translateX(${100 - presentationProgressWithEpsilonCorrection}%)`
        };
      case "from-top":
        return {
          transform: `translateY(${-100 + presentationProgress * 100}%)`
        };
      case "from-bottom":
        return {
          transform: `translateY(${100 - presentationProgressWithEpsilonCorrection}%)`
        };
      default:
        throw new Error(`Invalid direction: ${direction}`);
    }
  }, [presentationDirection, presentationProgress, direction]);
  const style = useMemo(() => {
    return {
      width: "100%",
      height: "100%",
      justifyContent: "center",
      alignItems: "center",
      ...directionStyle,
      ...presentationDirection === "entering" ? enterStyle : exitStyle
    };
  }, [directionStyle, enterStyle, exitStyle, presentationDirection]);
  return /* @__PURE__ */ jsx(AbsoluteFill, {
    style,
    children
  });
};
var slide = (props) => {
  return {
    component: SlidePresentation,
    props: props ?? {}
  };
};

// src/timings/linear-timing.ts
import { interpolate } from "remotion";
var linearTiming = (options) => {
  return {
    getDurationInFrames: () => {
      return options.durationInFrames;
    },
    getProgress: ({ frame }) => {
      return interpolate(frame, [0, options.durationInFrames], [0, 1], {
        easing: options.easing,
        extrapolateLeft: "clamp",
        extrapolateRight: "clamp"
      });
    }
  };
};
// src/timings/spring-timing.ts
import { measureSpring, spring } from "remotion";
var springTiming = (options = {}) => {
  return {
    getDurationInFrames: ({ fps }) => {
      if (options.durationInFrames) {
        return options.durationInFrames;
      }
      return measureSpring({
        config: options.config,
        threshold: options.durationRestThreshold,
        fps
      });
    },
    getProgress: ({ fps, frame }) => {
      const to = options.reverse ? 0 : 1;
      const from = options.reverse ? 1 : 0;
      return spring({
        fps,
        frame,
        to,
        from,
        config: options.config,
        durationInFrames: options.durationInFrames,
        durationRestThreshold: options.durationRestThreshold,
        reverse: options.reverse
      });
    }
  };
};
// src/TransitionSeries.tsx
import { Children, useMemo as useMemo3 } from "react";
import { Internals, Sequence, useCurrentFrame, useVideoConfig } from "remotion";
import { NoReactInternals as NoReactInternals2 } from "remotion/no-react";

// src/context.tsx
import React2, { useMemo as useMemo2 } from "react";
import { jsx as jsx2 } from "react/jsx-runtime";
var EnteringContext = React2.createContext(null);
var ExitingContext = React2.createContext(null);
var WrapInEnteringProgressContext = ({ presentationProgress, children }) => {
  const value = useMemo2(() => {
    return {
      enteringProgress: presentationProgress
    };
  }, [presentationProgress]);
  return /* @__PURE__ */ jsx2(EnteringContext.Provider, {
    value,
    children
  });
};
var WrapInExitingProgressContext = ({ presentationProgress, children }) => {
  const value = useMemo2(() => {
    return {
      exitingProgress: presentationProgress
    };
  }, [presentationProgress]);
  return /* @__PURE__ */ jsx2(ExitingContext.Provider, {
    value,
    children
  });
};

// src/flatten-children.ts
import React3 from "react";
var flattenChildren = (children) => {
  const childrenArray = React3.Children.toArray(children);
  return childrenArray.reduce((flatChildren, child) => {
    if (child.type === React3.Fragment) {
      return flatChildren.concat(flattenChildren(child.props.children));
    }
    flatChildren.push(child);
    return flatChildren;
  }, []);
};

// src/validate.ts
import { NoReactInternals } from "remotion/no-react";
var validateDurationInFrames = NoReactInternals.validateDurationInFrames;

// src/TransitionSeries.tsx
import { jsx as jsx3, Fragment } from "react/jsx-runtime";
var TransitionSeriesTransition = function(_props) {
  return null;
};
var SeriesOverlay = () => {
  return null;
};
var SeriesSequence = ({ children }) => {
  return /* @__PURE__ */ jsx3(Fragment, {
    children
  });
};
var TransitionSeriesChildren = ({
  children
}) => {
  const { fps } = useVideoConfig();
  const frame = useCurrentFrame();
  const childrenValue = useMemo3(() => {
    let transitionOffsets = 0;
    let startFrame = 0;
    const flattedChildren = flattenChildren(children);
    const overlayRenders = [];
    const sequenceDurations = [];
    let pendingOverlayValidation = false;
    const mainChildren = Children.map(flattedChildren, (child, i) => {
      const current = child;
      if (typeof current === "string") {
        if (current.trim() === "") {
          return null;
        }
        throw new TypeError(`The <TransitionSeries /> component only accepts a list of <TransitionSeries.Sequence /> components as its children, but you passed a string "${current}"`);
      }
      const hasPrev = flattedChildren[i - 1];
      const nextPrev = flattedChildren[i + 1];
      const prevIsTransition = typeof hasPrev === "string" || typeof hasPrev === "undefined" ? false : hasPrev.type === TransitionSeriesTransition;
      const prevIsOverlay = typeof hasPrev === "string" || typeof hasPrev === "undefined" ? false : hasPrev.type === SeriesOverlay;
      if (current.type === SeriesOverlay) {
        if (prevIsOverlay) {
          throw new TypeError(`A <TransitionSeries.Overlay /> component must not be followed by another <TransitionSeries.Overlay /> component (nth children = ${i - 1} and ${i})`);
        }
        if (prevIsTransition) {
          throw new TypeError(`A <TransitionSeries.Transition /> component must not be followed by a <TransitionSeries.Overlay /> component (nth children = ${i - 1} and ${i})`);
        }
        const nextIsTransition = typeof nextPrev === "string" || typeof nextPrev === "undefined" ? false : nextPrev.type === TransitionSeriesTransition;
        if (nextIsTransition) {
          throw new TypeError(`A <TransitionSeries.Overlay /> component must not be followed by a <TransitionSeries.Transition /> component (nth children = ${i} and ${i + 1})`);
        }
        const overlayProps = current.props;
        validateDurationInFrames(overlayProps.durationInFrames, {
          component: `of a <TransitionSeries.Overlay /> component`,
          allowFloats: false
        });
        const overlayOffset = overlayProps.offset ?? 0;
        if (Number.isNaN(overlayOffset)) {
          throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must not be NaN, but got NaN.`);
        }
        if (!Number.isFinite(overlayOffset)) {
          throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must be finite, but got ${overlayOffset}.`);
        }
        if (overlayOffset % 1 !== 0) {
          throw new TypeError(`The "offset" property of a <TransitionSeries.Overlay /> must be an integer, but got ${overlayOffset}.`);
        }
        const cutPoint = startFrame + transitionOffsets;
        const halfDuration = overlayProps.durationInFrames / 2;
        const overlayFrom = cutPoint - halfDuration + overlayOffset;
        if (overlayFrom < 0) {
          throw new TypeError(`A <TransitionSeries.Overlay /> extends before frame 0. The overlay starts at frame ${overlayFrom}. Reduce the duration or adjust the offset.`);
        }
        const prevSeqIdx = sequenceDurations.length - 1;
        if (prevSeqIdx >= 0) {
          const overlayStartInPrev = halfDuration - overlayOffset;
          if (overlayStartInPrev > sequenceDurations[prevSeqIdx]) {
            throw new TypeError(`A <TransitionSeries.Overlay /> extends beyond the previous sequence. The overlay needs ${overlayStartInPrev} frames before the cut, but the previous sequence is only ${sequenceDurations[prevSeqIdx]} frames long.`);
          }
        }
        pendingOverlayValidation = true;
        overlayRenders.push({
          cutPoint,
          overlayFrom,
          durationInFrames: overlayProps.durationInFrames,
          overlayOffset,
          halfDuration,
          children: overlayProps.children,
          index: i
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
      const prev = typeof hasPrev === "string" || typeof hasPrev === "undefined" ? null : hasPrev.type === TransitionSeriesTransition ? hasPrev : null;
      const next = typeof nextPrev === "string" || typeof nextPrev === "undefined" ? null : nextPrev.type === TransitionSeriesTransition ? nextPrev : null;
      const castedChildAgain = current;
      const debugInfo = `index = ${i}, duration = ${castedChildAgain.props.durationInFrames}`;
      if (!castedChildAgain?.props.children) {
        throw new TypeError(`A <TransitionSeries.Sequence /> component (${debugInfo}) was detected to not have any children. Delete it to fix this error.`);
      }
      const durationInFramesProp = castedChildAgain.props.durationInFrames;
      const {
        durationInFrames,
        children: _children,
        ...passedProps
      } = castedChildAgain.props;
      validateDurationInFrames(durationInFramesProp, {
        component: `of a <TransitionSeries.Sequence /> component`,
        allowFloats: true
      });
      const offset = castedChildAgain.props.offset ?? 0;
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
          fps
        });
        transitionOffsets -= duration;
      }
      let actualStartFrame = currentStartFrame + transitionOffsets;
      startFrame += durationInFramesProp + offset;
      if (actualStartFrame < 0) {
        startFrame -= actualStartFrame;
        actualStartFrame = 0;
      }
      sequenceDurations.push(durationInFramesProp);
      if (pendingOverlayValidation) {
        pendingOverlayValidation = false;
        const lastOverlay = overlayRenders[overlayRenders.length - 1];
        const framesAfterCut = lastOverlay.halfDuration + lastOverlay.overlayOffset;
        if (framesAfterCut > durationInFramesProp) {
          throw new TypeError(`A <TransitionSeries.Overlay /> extends beyond the next sequence. The overlay needs ${framesAfterCut} frames after the cut, but the next sequence is only ${durationInFramesProp} frames long.`);
        }
      }
      const nextProgress = next ? next.props.timing.getProgress({
        frame: frame - actualStartFrame - durationInFrames + next.props.timing.getDurationInFrames({ fps }),
        fps
      }) : null;
      const prevProgress = prev ? prev.props.timing.getProgress({
        frame: frame - actualStartFrame,
        fps
      }) : null;
      if (next && durationInFramesProp < next.props.timing.getDurationInFrames({ fps })) {
        throw new Error(`The duration of a <TransitionSeries.Sequence /> must not be shorter than the duration of the next <TransitionSeries.Transition />. The transition is ${next.props.timing.getDurationInFrames({ fps })} frames long, but the sequence is only ${durationInFramesProp} frames long (${debugInfo})`);
      }
      if (prev && durationInFramesProp < prev.props.timing.getDurationInFrames({ fps })) {
        throw new Error(`The duration of a <TransitionSeries.Sequence /> must not be shorter than the duration of the previous <TransitionSeries.Transition />. The transition is ${prev.props.timing.getDurationInFrames({ fps })} frames long, but the sequence is only ${durationInFramesProp} frames long (${debugInfo})`);
      }
      if (next && prev && nextProgress !== null && prevProgress !== null) {
        const nextPresentation = next.props.presentation ?? slide();
        const prevPresentation = prev.props.presentation ?? slide();
        const UppercaseNextPresentation = nextPresentation.component;
        const UppercasePrevPresentation = prevPresentation.component;
        return /* @__PURE__ */ jsx3(Sequence, {
          from: actualStartFrame,
          durationInFrames: durationInFramesProp,
          ...passedProps,
          name: passedProps.name || "<TS.Sequence>",
          children: /* @__PURE__ */ jsx3(UppercaseNextPresentation, {
            passedProps: nextPresentation.props ?? {},
            presentationDirection: "exiting",
            presentationProgress: nextProgress,
            presentationDurationInFrames: next.props.timing.getDurationInFrames({ fps }),
            children: /* @__PURE__ */ jsx3(WrapInExitingProgressContext, {
              presentationProgress: nextProgress,
              children: /* @__PURE__ */ jsx3(UppercasePrevPresentation, {
                passedProps: prevPresentation.props ?? {},
                presentationDirection: "entering",
                presentationProgress: prevProgress,
                presentationDurationInFrames: prev.props.timing.getDurationInFrames({ fps }),
                children: /* @__PURE__ */ jsx3(WrapInEnteringProgressContext, {
                  presentationProgress: prevProgress,
                  children: child
                })
              })
            })
          })
        }, i);
      }
      if (prevProgress !== null && prev) {
        const prevPresentation = prev.props.presentation ?? slide();
        const UppercasePrevPresentation = prevPresentation.component;
        return /* @__PURE__ */ jsx3(Sequence, {
          from: actualStartFrame,
          durationInFrames: durationInFramesProp,
          ...passedProps,
          name: passedProps.name || "<TS.Sequence>",
          children: /* @__PURE__ */ jsx3(UppercasePrevPresentation, {
            passedProps: prevPresentation.props ?? {},
            presentationDirection: "entering",
            presentationProgress: prevProgress,
            presentationDurationInFrames: prev.props.timing.getDurationInFrames({ fps }),
            children: /* @__PURE__ */ jsx3(WrapInEnteringProgressContext, {
              presentationProgress: prevProgress,
              children: child
            })
          })
        }, i);
      }
      if (nextProgress !== null && next) {
        const nextPresentation = next.props.presentation ?? slide();
        const UppercaseNextPresentation = nextPresentation.component;
        return /* @__PURE__ */ jsx3(Sequence, {
          from: actualStartFrame,
          durationInFrames: durationInFramesProp,
          ...passedProps,
          name: passedProps.name || "<TS.Sequence>",
          children: /* @__PURE__ */ jsx3(UppercaseNextPresentation, {
            passedProps: nextPresentation.props ?? {},
            presentationDirection: "exiting",
            presentationProgress: nextProgress,
            presentationDurationInFrames: next.props.timing.getDurationInFrames({ fps }),
            children: /* @__PURE__ */ jsx3(WrapInExitingProgressContext, {
              presentationProgress: nextProgress,
              children: child
            })
          })
        }, i);
      }
      return /* @__PURE__ */ jsx3(Sequence, {
        from: actualStartFrame,
        durationInFrames: durationInFramesProp,
        ...passedProps,
        name: passedProps.name || "<TS.Sequence>",
        children: child
      }, i);
    });
    const overlayElements = overlayRenders.map((overlayInfo) => {
      const info = overlayInfo;
      return /* @__PURE__ */ jsx3(Sequence, {
        from: Math.round(info.overlayFrom),
        durationInFrames: info.durationInFrames,
        name: "<TS.Overlay>",
        layout: "absolute-fill",
        children: info.children
      }, `overlay-${info.index}`);
    });
    return [...mainChildren || [], ...overlayElements];
  }, [children, fps, frame]);
  return /* @__PURE__ */ jsx3(Fragment, {
    children: childrenValue
  });
};
var TransitionSeries = ({ children, name, layout: passedLayout, ...otherProps }) => {
  const displayName = name ?? "<TransitionSeries>";
  const layout = passedLayout ?? "absolute-fill";
  if (NoReactInternals2.ENABLE_V5_BREAKING_CHANGES && layout !== "absolute-fill") {
    throw new TypeError(`The "layout" prop of <TransitionSeries /> is not supported anymore in v5. TransitionSeries' must be absolutely positioned.`);
  }
  return /* @__PURE__ */ jsx3(Sequence, {
    name: displayName,
    layout,
    ...otherProps,
    children: /* @__PURE__ */ jsx3(TransitionSeriesChildren, {
      children
    })
  });
};
TransitionSeries.Sequence = SeriesSequence;
TransitionSeries.Transition = TransitionSeriesTransition;
TransitionSeries.Overlay = SeriesOverlay;
Internals.addSequenceStackTraces(TransitionSeries);
Internals.addSequenceStackTraces(SeriesSequence);
Internals.addSequenceStackTraces(SeriesOverlay);
// src/use-transition-progress.ts
import React4 from "react";
var useTransitionProgress = () => {
  const entering = React4.useContext(EnteringContext);
  const exiting = React4.useContext(ExitingContext);
  if (!entering && !exiting) {
    return {
      isInTransitionSeries: false,
      entering: 1,
      exiting: 0
    };
  }
  return {
    isInTransitionSeries: true,
    entering: entering?.enteringProgress ?? 1,
    exiting: exiting?.exitingProgress ?? 0
  };
};
export {
  useTransitionProgress,
  springTiming,
  linearTiming,
  TransitionSeries
};
