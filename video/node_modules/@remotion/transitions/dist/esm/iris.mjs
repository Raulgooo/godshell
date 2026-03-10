// src/presentations/iris.tsx
import { translatePath } from "@remotion/paths";
import { makeCircle } from "@remotion/shapes";
import { useMemo, useState } from "react";
import { AbsoluteFill, random } from "remotion";
import { jsx, jsxs } from "react/jsx-runtime";
var IrisPresentation = ({ children, presentationDirection, presentationProgress, passedProps }) => {
  const maxRadius = Math.sqrt(passedProps.width ** 2 + passedProps.height ** 2) / 2;
  const minRadius = 0;
  const currentRadius = presentationDirection === "entering" ? minRadius + (maxRadius - minRadius) * presentationProgress : maxRadius - (maxRadius - minRadius) * presentationProgress;
  const { path } = makeCircle({
    radius: currentRadius
  });
  const translatedPath = translatePath(path, passedProps.width / 2 - currentRadius, passedProps.height / 2 - currentRadius);
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
var iris = (props) => {
  return { component: IrisPresentation, props: props ?? {} };
};
export {
  iris
};
