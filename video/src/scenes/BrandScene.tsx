import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";

export const BrandScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const text = "GODSHELL";
  const subtext = "eBPF-powered system intelligence for humans and AI.";

  const textSpring = spring({
    frame,
    fps,
    config: { damping: 100, stiffness: 40 },
  });

  const opacity = interpolate(frame, [0, 1 * fps], [0, 1]);
  const scale = interpolate(textSpring, [0, 1], [0.95, 1]);

  return (
    <div
      style={{
        flex: 1,
        backgroundColor: COLORS.bgDeep,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div style={{ opacity, transform: `scale(${scale})`, textAlign: "center" }}>
        <h1
          style={{
            fontFamily: FONT.sans,
            fontSize: 120,
            fontWeight: 800,
            color: COLORS.red,
            letterSpacing: 16,
            margin: 0,
            filter: `drop-shadow(0 0 30px rgba(255, 34, 34, 0.4))`,
          }}
        >
          {text.split("").map((char, i) => {
            const charOpacity = interpolate(
              frame - i * 3,
              [1 * fps, 2 * fps],
              [0, 1],
              { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
            );
            return (
              <span key={i} style={{ opacity: charOpacity }}>
                {char}
              </span>
            );
          })}
        </h1>
        <div 
           style={{ 
             height: 1, 
             width: 400, 
             background: `linear-gradient(90deg, transparent, ${COLORS.accentSubtle}, transparent)`,
             margin: "40px auto 20px"
           }} 
        />
        <p
          style={{
            fontFamily: FONT.mono,
            fontSize: 24,
            color: COLORS.textSecondary,
            letterSpacing: 2,
            opacity: interpolate(frame, [3 * fps, 4.5 * fps], [0, 1]),
          }}
        >
          {subtext}
        </p>
      </div>
    </div>
  );
};
