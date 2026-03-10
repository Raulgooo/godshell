// src/presentations/fade.tsx
import { useMemo } from "react";
import { AbsoluteFill } from "remotion";
import { jsx } from "react/jsx-runtime";
var FadePresentation = ({ children, presentationDirection, presentationProgress, passedProps }) => {
  const isEntering = presentationDirection === "entering";
  const style = useMemo(() => {
    return {
      opacity: isEntering ? presentationProgress : passedProps.shouldFadeOutExitingScene ? 1 - presentationProgress : 1,
      ...presentationDirection === "entering" ? passedProps.enterStyle : passedProps.exitStyle
    };
  }, [
    isEntering,
    passedProps.enterStyle,
    passedProps.exitStyle,
    passedProps.shouldFadeOutExitingScene,
    presentationDirection,
    presentationProgress
  ]);
  return /* @__PURE__ */ jsx(AbsoluteFill, {
    style,
    children
  });
};
var fade = (props) => {
  return {
    component: FadePresentation,
    props: props ?? {}
  };
};
export {
  fade
};
