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
export {
  slide
};
