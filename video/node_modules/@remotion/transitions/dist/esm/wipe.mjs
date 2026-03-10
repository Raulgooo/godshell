// src/presentations/wipe.tsx
import { useMemo, useState } from "react";
import { AbsoluteFill, random } from "remotion";
import { jsx, jsxs } from "react/jsx-runtime";
var makePathIn = (progress, direction) => {
  switch (direction) {
    case "from-left":
      return `
M 0 0
L ${progress} 0
L ${progress} 1
L 0 1
Z`;
    case "from-top-left":
      return `
M 0 0
L ${progress * 2} 0
L 0 ${progress * 2}
Z`;
    case "from-top":
      return `
M 0 0
L 1 0
L 1 ${progress}
L 0 ${progress}
Z`;
    case "from-top-right":
      return `
M 1 0
L ${1 - progress * 2} 0
L 1 ${progress * 2}
Z`;
    case "from-right":
      return `
M 1 0
L 1 1
L ${1 - progress} 1
L ${1 - progress} 0
Z`;
    case "from-bottom-right":
      return `
M 1 1
L ${1 - progress * 2} 1
L 1 ${1 - progress * 2}
Z`;
    case "from-bottom":
      return `
M 0 1
L 1 1
L 1 ${1 - progress}
L 0 ${1 - progress}
Z`;
    case "from-bottom-left":
      return `
M 0 1
L 0 ${1 - progress * 2}
L ${progress * 2} 1
Z`;
    default:
      throw new Error(`Unknown direction ${JSON.stringify(direction)}`);
  }
};
var makePathOut = (progress, direction) => {
  switch (direction) {
    case "from-left":
      return `
M 1 1
L ${1 - progress} 1
L ${1 - progress} 0
L 1 0
Z`;
    case "from-top-left":
      return `
M 1 1
L ${1 - 2 * progress} 1
L 1 ${1 - 2 * progress}
Z`;
    case "from-top":
      return `
M 1 1
L 0 1
L 0 ${1 - progress}
L 1 ${1 - progress}
Z`;
    case "from-top-right":
      return `
M 0 1
L ${progress * 2} 1
L 0 ${1 - progress * 2}
Z`;
    case "from-right":
      return `
M 0 0
L ${progress} 0
L ${progress} 1
L 0 1
Z`;
    case "from-bottom-right":
      return `
M 0 0
L ${progress * 2} 0
L 0 ${progress * 2}
Z`;
    case "from-bottom":
      return `
M 1 0
L 0 0
L 0 ${progress}
L 1 ${progress}
Z`;
    case "from-bottom-left":
      return `
M 1 0
L ${1 - progress * 2} 0
L 1 ${progress * 2}
Z`;
    default:
      throw new Error(`Unknown direction ${JSON.stringify(direction)}`);
  }
};
var WipePresentation = ({
  children,
  presentationProgress,
  presentationDirection,
  passedProps: {
    direction = "from-left",
    innerEnterStyle,
    innerExitStyle,
    outerEnterStyle,
    outerExitStyle
  }
}) => {
  const [clipId] = useState(() => String(random(null)));
  const progressInDirection = presentationDirection === "entering" ? presentationProgress : 1 - presentationProgress;
  const path = presentationDirection === "entering" ? makePathIn(progressInDirection, direction) : makePathOut(progressInDirection, direction);
  const style = useMemo(() => {
    return {
      width: "100%",
      height: "100%",
      justifyContent: "center",
      alignItems: "center",
      clipPath: `url(#${clipId})`,
      ...presentationDirection === "entering" ? innerEnterStyle : innerExitStyle
    };
  }, [clipId, innerEnterStyle, innerExitStyle, presentationDirection]);
  const outerStyle = useMemo(() => {
    return presentationDirection === "entering" ? outerEnterStyle : outerExitStyle;
  }, [outerEnterStyle, outerExitStyle, presentationDirection]);
  const svgStyle = useMemo(() => {
    return {
      width: "100%",
      height: "100%",
      pointerEvents: "none"
    };
  }, []);
  return /* @__PURE__ */ jsxs(AbsoluteFill, {
    style: outerStyle,
    children: [
      /* @__PURE__ */ jsx(AbsoluteFill, {
        style,
        children
      }),
      /* @__PURE__ */ jsx(AbsoluteFill, {
        children: /* @__PURE__ */ jsx("svg", {
          viewBox: "0 0 1 1",
          style: svgStyle,
          children: /* @__PURE__ */ jsx("defs", {
            children: /* @__PURE__ */ jsx("clipPath", {
              id: clipId,
              clipPathUnits: "objectBoundingBox",
              children: /* @__PURE__ */ jsx("path", {
                d: path,
                fill: "black"
              })
            })
          })
        })
      })
    ]
  });
};
var wipe = (props) => {
  return {
    component: WipePresentation,
    props: props ?? {}
  };
};
export {
  wipe
};
