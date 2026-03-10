import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  spring,
} from "remotion";
import { Shield, Wrench, MemoryStick } from "lucide-react";
import { COLORS, FONT, SPACING } from "../styles";

const DEMOS = [
  {
    Icon: Shield,
    title: "Catch a Reverse Shell",
    desc: "Detect C2 connections the moment they form",
    tools: ["gonetwork_state", "trace", "goread_environ"],
  },
  {
    Icon: Wrench,
    title: "Debug a Broken Build",
    desc: "See exactly which subprocess crashed and why",
    tools: ["inspect", "family", "read_file"],
  },
  {
    Icon: MemoryStick,
    title: "Find the Memory Leak",
    desc: "Compare heap maps across processes in real-time",
    tools: ["get_maps", "get_libraries", "trace"],
  },
];

export const DemoCardsScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const titleOpacity = interpolate(frame, [0, 0.5 * fps], [0, 1], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });
  const titleY = interpolate(frame, [0, 0.5 * fps], [10, 0], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.bg }}>
      {/* Title */}
      <div
        style={{
          position: "absolute",
          top: 96,
          left: 0,
          right: 0,
          textAlign: "center",
          opacity: titleOpacity,
          transform: `translateY(${titleY}px)`,
        }}
      >
        <div
          style={{
            fontFamily: FONT.sans,
            fontSize: 36,
            color: COLORS.text,
            fontWeight: 600,
            letterSpacing: -0.5,
          }}
        >
          Security. Debugging. Observability.
        </div>
        <div
          style={{
            fontFamily: FONT.sans,
            fontSize: 17,
            fontWeight: 300,
            color: COLORS.textDim,
            marginTop: 10,
            letterSpacing: 0.5,
          }}
        >
          One shell. Every use case.
        </div>
      </div>

      {/* Cards */}
      <div
        style={{
          position: "absolute",
          top: 0,
          bottom: 0,
          left: 0,
          right: 0,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          gap: 28,
          padding: "0 140px",
          marginTop: 40,
        }}
      >
        {DEMOS.map((demo, i) => {
          const delay = 0.8 + i * 0.9;
          const cardProgress = spring({
            frame: frame - delay * fps,
            fps,
            config: { damping: 16, stiffness: 90 },
            durationInFrames: Math.round(1 * fps),
          });
          const cardOpacity = interpolate(
            frame - delay * fps,
            [0, 0.2 * fps],
            [0, 1],
            { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
          );
          const cardY = interpolate(cardProgress, [0, 1], [36, 0]);

          const IconComponent = demo.Icon;

          return (
            <div
              key={i}
              style={{
                opacity: cardOpacity,
                transform: `translateY(${cardY}px)`,
                width: 400,
                borderRadius: 16,
                background: `linear-gradient(160deg, #1C1C1C 0%, #141414 100%)`,
                border: `1px solid ${COLORS.border}`,
                overflow: "hidden",
                boxShadow: "0 12px 40px rgba(0,0,0,0.5)",
              }}
            >
              {/* Red accent bar */}
              <div
                style={{
                  height: 2,
                  background: `linear-gradient(90deg, ${COLORS.accent} 0%, ${COLORS.accentMuted} 100%)`,
                }}
              />

              <div
                style={{
                  padding: "32px 28px 28px",
                  display: "flex",
                  flexDirection: "column",
                  gap: 16,
                }}
              >
                {/* Icon container */}
                <div
                  style={{
                    width: 46,
                    height: 46,
                    borderRadius: 12,
                    background: COLORS.bgSubtle,
                    border: `1px solid ${COLORS.border}`,
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                  }}
                >
                  <IconComponent
                    size={22}
                    color={COLORS.accent}
                    strokeWidth={1.5}
                  />
                </div>

                <div
                  style={{
                    fontFamily: FONT.sans,
                    fontSize: 22,
                    color: COLORS.text,
                    fontWeight: 600,
                    lineHeight: 1.3,
                    letterSpacing: -0.3,
                  }}
                >
                  {demo.title}
                </div>

                <div
                  style={{
                    fontFamily: FONT.sans,
                    fontSize: 15,
                    color: COLORS.textSecondary,
                    lineHeight: 1.65,
                    fontWeight: 300,
                  }}
                >
                  {demo.desc}
                </div>

                {/* Tool badges */}
                <div
                  style={{
                    display: "flex",
                    flexWrap: "wrap",
                    gap: 6,
                    marginTop: 2,
                  }}
                >
                  {demo.tools.map((tool) => (
                    <span
                      key={tool}
                      style={{
                        fontFamily: FONT.mono,
                        fontSize: 11,
                        color: COLORS.textDim,
                        padding: "4px 10px",
                        borderRadius: 6,
                        border: `1px solid ${COLORS.borderSubtle}`,
                        background: COLORS.bgSubtle,
                      }}
                    >
                      {tool}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
    </AbsoluteFill>
  );
};
