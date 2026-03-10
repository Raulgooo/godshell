// src/presentations/flip.tsx
import { useMemo } from "react";
import { AbsoluteFill, interpolate } from "remotion";
import { jsx } from "react/jsx-runtime";
var Flip = ({
  children,
  presentationDirection,
  presentationProgress,
  passedProps: {
    direction = "from-left",
    perspective = 1000,
    innerEnterStyle,
    innerExitStyle,
    outerEnterStyle,
    outerExitStyle
  }
}) => {
  const style = useMemo(() => {
    const startRotationEntering = direction === "from-right" || direction === "from-top" ? 180 : -180;
    const endRotationEntering = direction === "from-right" || direction === "from-top" ? -180 : 180;
    const rotation = presentationDirection === "entering" ? interpolate(presentationProgress, [0, 1], [startRotationEntering, 0]) : interpolate(presentationProgress, [0, 1], [0, endRotationEntering]);
    const rotateProperty = direction === "from-top" || direction === "from-bottom" ? "rotateX" : "rotateY";
    return {
      width: "100%",
      height: "100%",
      transform: `${rotateProperty}(${rotation}deg)`,
      backfaceVisibility: "hidden",
      WebkitBackfaceVisibility: "hidden",
      ...presentationDirection === "entering" ? innerEnterStyle : innerExitStyle
    };
  }, [
    direction,
    innerEnterStyle,
    innerExitStyle,
    presentationDirection,
    presentationProgress
  ]);
  const outer = useMemo(() => {
    return {
      perspective,
      transformStyle: "preserve-3d",
      ...presentationDirection === "entering" ? outerEnterStyle : outerExitStyle
    };
  }, [outerEnterStyle, outerExitStyle, perspective, presentationDirection]);
  return /* @__PURE__ */ jsx(AbsoluteFill, {
    style: outer,
    children: /* @__PURE__ */ jsx(AbsoluteFill, {
      style,
      children
    })
  });
};
var flip = (props) => {
  return { component: Flip, props: props ?? {} };
};
export {
  flip
};
