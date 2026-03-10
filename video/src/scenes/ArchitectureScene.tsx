import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
  spring,
} from "remotion";
import { Zap, Dna, RefreshCw, Brain, MessageSquare, User } from "lucide-react";
import { COLORS, FONT, SPACING } from "../styles";

const NODES = [
  { label: "Kernel", sub: "syscalls", Icon: Zap, color: COLORS.accent },
  { label: "eBPF", sub: "zero overhead", Icon: Dna, color: COLORS.accent },
  { label: "Ring Buffer", sub: "lock-free", Icon: RefreshCw, color: COLORS.accent },
  { label: "Context Engine", sub: "process tree", Icon: Brain, color: COLORS.accent },
  { label: "LLM", sub: "14 tools", Icon: MessageSquare, color: COLORS.accent },
  { label: "You", sub: "just ask", Icon: User, color: COLORS.text },
];

export const ArchitectureScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const titleOpacity = interpolate(frame, [0, 0.7 * fps], [0, 1], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });
  const titleY = interpolate(frame, [0, 0.7 * fps], [14, 0], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  const flowProgress = interpolate(
    frame,
    [2 * fps, 6 * fps],
    [0, 1],
    { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
  );

  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.bg }}>
      {/* Subtle dot grid */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          backgroundImage: `radial-gradient(${COLORS.border} 1px, transparent 1px)`,
          backgroundSize: "40px 40px",
          opacity: 0.25,
        }}
      />

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
            fontSize: 34,
            color: COLORS.text,
            fontWeight: 600,
            letterSpacing: -0.5,
          }}
        >
          Zero probing. Kernel-native observability.
        </div>
        <div
          style={{
            fontFamily: FONT.mono,
            fontSize: 12,
            color: COLORS.textDim,
            marginTop: 12,
            letterSpacing: 4,
            textTransform: "uppercase",
          }}
        >
          from syscall to insight in milliseconds
        </div>
      </div>

      {/* Architecture pipeline */}
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
          gap: 12,
          padding: "0 100px",
        }}
      >
        {NODES.map((node, i) => {
          const delay = 0.5 + i * 0.55;
          const nodeProgress = spring({
            frame: frame - delay * fps,
            fps,
            config: { damping: 200 },
            durationInFrames: Math.round(0.7 * fps),
          });
          const nodeOpacity = interpolate(
            frame - delay * fps,
            [0, 0.3 * fps],
            [0, 1],
            { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
          );
          const nodeY = interpolate(nodeProgress, [0, 1], [24, 0]);

          const lineDelay = delay + 0.3;
          const lineOpacity = interpolate(
            frame - lineDelay * fps,
            [0, 0.25 * fps],
            [0, 1],
            { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
          );

          const flowThreshold = i / NODES.length;
          const isLit = flowProgress > flowThreshold;

          // Pulsing dot on connector
          const dotProgress = isLit
            ? (flowProgress - flowThreshold) / (1 / NODES.length)
            : 0;

          const IconComponent = node.Icon;

          return (
            <React.Fragment key={i}>
              <div
                style={{
                  opacity: nodeOpacity,
                  transform: `translateY(${nodeY}px)`,
                  display: "flex",
                  flexDirection: "column",
                  alignItems: "center",
                  gap: 14,
                  flex: 1,
                }}
              >
                {/* Icon card */}
                <div
                  style={{
                    width: 88,
                    height: 88,
                    borderRadius: 18,
                    background: isLit
                      ? `linear-gradient(145deg, #1E1E1E, #131313)`
                      : `linear-gradient(145deg, #181818, #111111)`,
                    border: `1px solid ${isLit ? COLORS.accent + "50" : COLORS.border}`,
                    display: "flex",
                    justifyContent: "center",
                    alignItems: "center",
                    boxShadow: isLit
                      ? `0 0 24px rgba(220,38,38,0.2), inset 0 1px 0 rgba(255,255,255,0.04)`
                      : `inset 0 1px 0 rgba(255,255,255,0.03)`,
                  }}
                >
                  <IconComponent
                    size={32}
                    color={isLit ? COLORS.accent : COLORS.textDim}
                    strokeWidth={1.5}
                  />
                </div>
                <div
                  style={{
                    fontFamily: FONT.sans,
                    fontSize: 14,
                    color: COLORS.text,
                    fontWeight: 500,
                    letterSpacing: 0.1,
                  }}
                >
                  {node.label}
                </div>
                <div
                  style={{
                    fontFamily: FONT.mono,
                    fontSize: 11,
                    color: COLORS.textDim,
                    letterSpacing: 0.3,
                  }}
                >
                  {node.sub}
                </div>
              </div>

              {/* Connector */}
              {i < NODES.length - 1 && (
                <div
                  style={{
                    opacity: lineOpacity,
                    display: "flex",
                    alignItems: "center",
                    marginTop: -42,
                    position: "relative",
                  }}
                >
                  {/* Line */}
                  <div
                    style={{
                      width: 32,
                      height: 1,
                      background: isLit ? COLORS.accent : COLORS.border,
                      boxShadow: isLit ? `0 0 8px ${COLORS.accent}50` : "none",
                    }}
                  />
                  {/* Traveling dot */}
                  {isLit && (
                    <div
                      style={{
                        position: "absolute",
                        left: Math.min(dotProgress * 30, 28),
                        top: -2,
                        width: 5,
                        height: 5,
                        borderRadius: "50%",
                        background: COLORS.accent,
                        boxShadow: `0 0 6px ${COLORS.accent}`,
                        opacity: interpolate(dotProgress, [0, 0.2, 0.8, 1], [0, 1, 1, 0], {
                          extrapolateRight: "clamp",
                          extrapolateLeft: "clamp",
                        }),
                      }}
                    />
                  )}
                  {/* Chevron */}
                  <svg width="7" height="10" viewBox="0 0 7 10" style={{ marginLeft: -1, opacity: isLit ? 0.6 : 0.3 }}>
                    <path d="M1 1 L5 5 L1 9" stroke={isLit ? COLORS.accent : COLORS.textDim} strokeWidth="1.5" fill="none" />
                  </svg>
                </div>
              )}
            </React.Fragment>
          );
        })}
      </div>

      {/* Tracepoint badges */}
      <div
        style={{
          position: "absolute",
          bottom: 86,
          left: 0,
          right: 0,
          textAlign: "center",
          opacity: interpolate(frame, [4 * fps, 5.2 * fps], [0, 1], {
            extrapolateRight: "clamp",
            extrapolateLeft: "clamp",
          }),
        }}
      >
        <div style={{ display: "flex", justifyContent: "center", gap: 10 }}>
          {["sys_enter_execve", "sys_enter_openat", "sys_enter_connect", "sched_process_exit"].map((tp) => (
            <span
              key={tp}
              style={{
                fontFamily: FONT.mono,
                fontSize: 13,
                color: COLORS.textDim,
                padding: "6px 14px",
                borderRadius: 8,
                border: `1px solid ${COLORS.border}`,
                background: COLORS.surface,
              }}
            >
              {tp}
            </span>
          ))}
        </div>
      </div>

      {/* Letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
    </AbsoluteFill>
  );
};
