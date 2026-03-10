import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
} from "remotion";
import { COLORS, FONT, SPACING } from "../styles";

const CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*";
const TARGET = "Your AI terminal is blind.";

const seededRandom = (seed: number) => {
  const x = Math.sin(seed * 9301 + 49297) * 49999;
  return x - Math.floor(x);
};

export const HookScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const scrambleStart = 0.8 * fps;
  const scrambleEnd = 2.5 * fps;
  const scrambleDuration = scrambleEnd - scrambleStart;

  const sceneOpacity = interpolate(frame, [0, 0.6 * fps], [0, 1], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  const getDisplayText = () => {
    if (frame < scrambleStart) return "";
    const progress = Math.min(1, (frame - scrambleStart) / scrambleDuration);
    const resolvedChars = Math.floor(progress * TARGET.length);
    return TARGET.split("")
      .map((char, i) => {
        if (i < resolvedChars) return char;
        if (char === " ") return " ";
        const randIdx = Math.floor(
          seededRandom(frame * 100 + i * 7) * CHARS.length
        );
        return CHARS[randIdx];
      })
      .join("");
  };

  const glowIntensity = interpolate(
    frame,
    [scrambleEnd, scrambleEnd + 0.5 * fps],
    [0, 1],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Ambient glow orb — appears after scramble resolves
  const orbOpacity = interpolate(
    frame,
    [scrambleEnd, scrambleEnd + 1 * fps],
    [0, 0.5],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  // Horizontal ambient light streaks — slow drift upward
  const streaks = Array.from({ length: 5 }, (_, i) => {
    const baseY = 200 + seededRandom(i * 11 + 1) * 680;
    const width = 300 + seededRandom(i * 11 + 2) * 500;
    const speed = 0.08 + seededRandom(i * 11 + 3) * 0.12;
    const y = ((baseY - frame * speed) + 1080) % 1080;
    const opacity = (0.03 + seededRandom(i * 11 + 4) * 0.05) *
      interpolate(frame, [0.5 * fps, 1.5 * fps], [0, 1], {
        extrapolateRight: "clamp",
        extrapolateLeft: "clamp",
      });
    return { y, width, opacity };
  });

  return (
    <AbsoluteFill
      style={{
        backgroundColor: COLORS.bgDeep,
        opacity: sceneOpacity,
      }}
    >
      {/* Ambient glow orb — centered, appears after scramble */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          top: "50%",
          width: 800,
          height: 400,
          marginLeft: -400,
          marginTop: -200,
          borderRadius: "50%",
          background: `radial-gradient(ellipse, ${COLORS.accent}18 0%, transparent 65%)`,
          opacity: orbOpacity,
          filter: "blur(60px)",
        }}
      />

      {/* Horizontal ambient streaks */}
      {streaks.map((s, i) => (
        <div
          key={i}
          style={{
            position: "absolute",
            left: "50%",
            top: s.y,
            width: s.width,
            height: 1,
            marginLeft: -(s.width / 2),
            background: `linear-gradient(90deg, transparent 0%, ${COLORS.accent}50 50%, transparent 100%)`,
            opacity: s.opacity,
          }}
        />
      ))}

      {/* Top vignette */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: 260, background: `linear-gradient(180deg, ${COLORS.bgDeep} 0%, transparent 100%)` }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: 260, background: `linear-gradient(0deg, ${COLORS.bgDeep} 0%, transparent 100%)` }} />

      {/* Main text */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <div
          style={{
            fontFamily: FONT.mono,
            fontSize: 88,
            fontWeight: 600,
            color: COLORS.text,
            letterSpacing: -0.5,
            textShadow: glowIntensity > 0
              ? `0 0 ${40 * glowIntensity}px ${COLORS.accent}25`
              : "none",
          }}
        >
          {getDisplayText()}
        </div>
      </div>

      {/* Cinematic letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bgDeep }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bgDeep }} />
    </AbsoluteFill>
  );
};
