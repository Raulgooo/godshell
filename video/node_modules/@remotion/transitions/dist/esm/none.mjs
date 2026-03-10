// src/presentations/none.tsx
import { useMemo } from "react";
import { AbsoluteFill } from "remotion";
import { jsx } from "react/jsx-runtime";
var NonePresentation = ({ children, presentationDirection, passedProps }) => {
  const style = useMemo(() => {
    return {
      ...presentationDirection === "entering" ? passedProps.enterStyle : passedProps.exitStyle
    };
  }, [passedProps.enterStyle, passedProps.exitStyle, presentationDirection]);
  return /* @__PURE__ */ jsx(AbsoluteFill, {
    style,
    children
  });
};
var none = (props) => {
  return {
    component: NonePresentation,
    props: props ?? {}
  };
};
export {
  none
};
