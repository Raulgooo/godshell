import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  spring,
} from "remotion";
import { COLORS, FONT, SPACING } from "../styles";

export const SolutionScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // "What if it already knew?" — ultra-thin, clean entrance
  const line1Progress = spring({
    frame,
    fps,
    config: { damping: 200 },
    durationInFrames: Math.round(1 * fps),
  });

  // GODshell reveal at 1.5s — dramatic spring
  const godshellDelay = 1.5 * fps;
  const godshellProgress = spring({
    frame: frame - godshellDelay,
    fps,
    config: { damping: 12, stiffness: 50 },
    durationInFrames: Math.round(1.8 * fps),
  });
  const godshellScale = interpolate(godshellProgress, [0, 1], [0.65, 1]);
  const godshellOpacity = interpolate(
    frame - godshellDelay,
    [0, 0.4 * fps],
    [0, 1],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Dramatic flash burst on reveal
  const burstOpacity = interpolate(
    frame - godshellDelay,
    [0, 0.06 * fps, 0.6 * fps],
    [0, 0.6, 0],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Tagline at 3s
  const taglineDelay = 3 * fps;
  const taglineOpacity = interpolate(
    frame - taglineDelay,
    [0, 0.8 * fps],
    [0, 1],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Animated rule above tagline
  const ruleWidth = interpolate(
    frame - taglineDelay,
    [0, 0.6 * fps],
    [0, 120],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Red glow behind logo — slow bloom
  const glowScale = interpolate(
    frame - godshellDelay,
    [0, 2.5 * fps],
    [0.2, 1.8],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );
  const glowOpacity = interpolate(
    frame - godshellDelay,
    [0, 1.2 * fps],
    [0, 0.4],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.bg }}>
      {/* Red glow orb */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          top: "50%",
          width: 700,
          height: 700,
          marginLeft: -350,
          marginTop: -350,
          borderRadius: "50%",
          background: `radial-gradient(circle, ${COLORS.accent}35 0%, transparent 60%)`,
          opacity: glowOpacity,
          transform: `scale(${glowScale})`,
          filter: "blur(100px)",
        }}
      />

      {/* Flash burst */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: `radial-gradient(circle at 50% 50%, ${COLORS.accent}55, transparent 35%)`,
          opacity: burstOpacity,
          mixBlendMode: "screen",
        }}
      />

      {/* Content */}
      <div
        style={{
          position: "relative",
          zIndex: 1,
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          gap: 20,
        }}
      >
        {/* Line 1 — ultra-thin */}
        <div
          style={{
            opacity: line1Progress,
            transform: `translateY(${interpolate(line1Progress, [0, 1], [18, 0])}px)`,
            fontFamily: FONT.sans,
            fontSize: 44,
            color: COLORS.textSecondary,
            fontWeight: 200,
            letterSpacing: 1,
          }}
        >
          What if it already knew?
        </div>

        {/* GODshell — gradient text */}
        <div
          style={{
            opacity: godshellOpacity,
            transform: `scale(${godshellScale})`,
            fontFamily: FONT.mono,
            fontSize: 144,
            fontWeight: 800,
            letterSpacing: -4,
            background: `linear-gradient(135deg, ${COLORS.text} 0%, ${COLORS.accent} 55%, ${COLORS.accentMuted} 100%)`,
            WebkitBackgroundClip: "text",
            WebkitTextFillColor: "transparent",
            filter: `drop-shadow(0 0 60px ${COLORS.accent}25)`,
          }}
        >
          GODshell
        </div>

        {/* Animated rule + tagline */}
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            gap: 16,
            opacity: taglineOpacity,
          }}
        >
          {/* Horizontal rule */}
          <div
            style={{
              height: 1,
              width: ruleWidth,
              background: `linear-gradient(90deg, transparent, ${COLORS.border}, transparent)`,
            }}
          />
          <div
            style={{
              fontFamily: FONT.sans,
              fontSize: 24,
              color: COLORS.textSecondary,
              fontWeight: 300,
              letterSpacing: 2,
            }}
          >
            Fed from the kernel. Not probing the OS.
          </div>
        </div>
      </div>

      {/* Letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
    </AbsoluteFill>
  );
};
