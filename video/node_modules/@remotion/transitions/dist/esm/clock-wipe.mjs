// src/presentations/clock-wipe.tsx
import { translatePath } from "@remotion/paths";
import { makePie } from "@remotion/shapes";
import { useMemo, useState } from "react";
import { AbsoluteFill, random } from "remotion";
import { jsx, jsxs } from "react/jsx-runtime";
var ClockWipePresentation = ({ children, presentationDirection, presentationProgress, passedProps }) => {
  const finishedRadius = Math.sqrt(passedProps.width ** 2 + passedProps.height ** 2) / 2;
  const { path } = makePie({
    radius: finishedRadius,
    progress: presentationProgress
  });
  const translatedPath = translatePath(path, -(finishedRadius * 2 - passedProps.width) / 2, -(finishedRadius * 2 - passedProps.height) / 2);
  const [clipId] = useState(() => String(random(null)));
  const style = useMemo(() => {
    return {
      width: "100%",
      height: "100%",
      clipPath: presentationDirection === "exiting" ? undefined : `url(#${clipId})`,
      ...presentationDirection === "entering" ? passedProps.innerEnterStyle : passedProps.innerExitStyle
    };
  }, [
    clipId,
    passedProps.innerEnterStyle,
    passedProps.innerExitStyle,
    presentationDirection
  ]);
  const outerStyle = useMemo(() => {
    return presentationDirection === "entering" ? passedProps.outerEnterStyle : passedProps.outerExitStyle;
  }, [
    passedProps.outerEnterStyle,
    passedProps.outerExitStyle,
    presentationDirection
  ]);
  return /* @__PURE__ */ jsxs(AbsoluteFill, {
    style: outerStyle,
    children: [
      /* @__PURE__ */ jsx(AbsoluteFill, {
        style,
        children
      }),
      presentationDirection === "exiting" ? null : /* @__PURE__ */ jsx(AbsoluteFill, {
        children: /* @__PURE__ */ jsx("svg", {
          children: /* @__PURE__ */ jsx("defs", {
            children: /* @__PURE__ */ jsx("clipPath", {
              id: clipId,
              children: /* @__PURE__ */ jsx("path", {
                d: translatedPath,
                fill: "black"
              })
            })
          })
        })
      })
    ]
  });
};
var clockWipe = (props) => {
  return { component: ClockWipePresentation, props: props ?? {} };
};
export {
  clockWipe
};
