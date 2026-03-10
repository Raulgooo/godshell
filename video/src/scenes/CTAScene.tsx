import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  spring,
} from "remotion";
import { COLORS, FONT, SPACING } from "../styles";

// Deterministic noise grain via inline SVG filter
const GrainOverlay: React.FC<{ opacity: number }> = ({ opacity }) => (
  <svg
    style={{
      position: "absolute",
      inset: 0,
      width: "100%",
      height: "100%",
      opacity,
      mixBlendMode: "overlay",
      pointerEvents: "none",
    }}
  >
    <filter id="grain">
      <feTurbulence
        type="fractalNoise"
        baseFrequency="0.65"
        numOctaves="3"
        stitchTiles="stitch"
      />
      <feColorMatrix type="saturate" values="0" />
    </filter>
    <rect width="100%" height="100%" filter="url(#grain)" />
  </svg>
);

export const CTAScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Logo — dramatic spring from small scale
  const logoProgress = spring({
    frame,
    fps,
    config: { damping: 12, stiffness: 45 },
    durationInFrames: Math.round(1.8 * fps),
  });
  const logoScale = interpolate(logoProgress, [0, 1], [0.4, 1]);
  const logoOpacity = interpolate(frame, [0, 0.5 * fps], [0, 1], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  // Tagline at 1.4s
  const taglineDelay = 1.4 * fps;
  const taglineProgress = spring({
    frame: frame - taglineDelay,
    fps,
    config: { damping: 200 },
    durationInFrames: Math.round(0.8 * fps),
  });

  // Link at 2.4s
  const linkDelay = 2.4 * fps;
  const linkProgress = spring({
    frame: frame - linkDelay,
    fps,
    config: { damping: 200 },
    durationInFrames: Math.round(0.6 * fps),
  });

  // Grain opacity — fades in slowly
  const grainOpacity = interpolate(frame, [0.5 * fps, 2 * fps], [0, 0.025], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  // Red glow — slow bloom
  const glowScale = interpolate(frame, [0, 3 * fps], [0.4, 1.8], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });
  const glowOpacity = interpolate(frame, [0, 2 * fps], [0, 0.35], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.bgDeep }}>
      {/* Red glow */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          top: "48%",
          width: 800,
          height: 800,
          marginLeft: -400,
          marginTop: -400,
          borderRadius: "50%",
          background: `radial-gradient(circle, ${COLORS.accent}28 0%, transparent 50%)`,
          opacity: glowOpacity,
          transform: `scale(${glowScale})`,
          filter: "blur(120px)",
        }}
      />

      {/* Film grain overlay */}
      <GrainOverlay opacity={grainOpacity} />

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
          gap: 24,
        }}
      >
        {/* Logo */}
        <div
          style={{
            opacity: logoOpacity,
            transform: `scale(${logoScale})`,
            fontFamily: FONT.mono,
            fontSize: 144,
            fontWeight: 800,
            letterSpacing: -5,
            background: `linear-gradient(135deg, #ffffff 0%, ${COLORS.accent} 50%, ${COLORS.accentMuted} 100%)`,
            WebkitBackgroundClip: "text",
            WebkitTextFillColor: "transparent",
            filter: `drop-shadow(0 0 50px ${COLORS.accent}20)`,
          }}
        >
          GODshell
        </div>

        {/* Tagline — ultra-thin tracked out */}
        <div
          style={{
            opacity: taglineProgress,
            transform: `translateY(${interpolate(taglineProgress, [0, 1], [12, 0])}px)`,
            fontFamily: FONT.sans,
            fontSize: 28,
            color: COLORS.textSecondary,
            fontWeight: 200,
            letterSpacing: 4,
          }}
        >
          A shell that already knows.
        </div>

        {/* GitHub — fully pill-shaped */}
        <div
          style={{
            opacity: linkProgress,
            transform: `translateY(${interpolate(linkProgress, [0, 1], [10, 0])}px)`,
            fontFamily: FONT.mono,
            fontSize: 16,
            color: COLORS.textDim,
            padding: "14px 36px",
            borderRadius: 100,
            border: `1px solid rgba(255,255,255,0.1)`,
            background: "rgba(255,255,255,0.03)",
            letterSpacing: 0.5,
          }}
        >
          github.com/raulcodes/godshell
        </div>
      </div>

      {/* Letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bgDeep }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bgDeep }} />
    </AbsoluteFill>
  );
};
