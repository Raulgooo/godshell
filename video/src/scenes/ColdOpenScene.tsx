import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate } from "remotion";
import { COLORS } from "../styles";

export const ColdOpenScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Cursor blink
  const cursorOpacity = Math.floor((frame / fps) * 2) % 2 === 0 ? 1 : 0;

  return (
    <div
      style={{
        flex: 1,
        backgroundColor: COLORS.bgDeep,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div
        style={{
          width: 12,
          height: 24,
          backgroundColor: COLORS.red,
          opacity: cursorOpacity,
          boxShadow: `0 0 10px ${COLORS.red}`,
        }}
      />
    </div>
  );
};
