import React from "react";
import {
  AbsoluteFill,
  useCurrentFrame,
  useVideoConfig,
  interpolate,
} from "remotion";
import { Activity, ChevronRight, Network, GitFork, AlertTriangle } from "lucide-react";
import { Terminal } from "../components/Terminal";
import { Typewriter } from "../components/Typewriter";
import { COLORS, FONT, SPACING } from "../styles";

const RESPONSE_LINES = [
  { icon: Activity, text: "Scanning system snapshot...", color: COLORS.textDim, bold: false },
  { icon: null, text: "", color: "", bold: false },
  { icon: AlertTriangle, text: "python3 (PID 8421) — 94.2% CPU", color: COLORS.accent, bold: true },
  { icon: ChevronRight, text: "cmdline: python3 crypto_worker.py --threads=8", color: COLORS.text, bold: false },
  { icon: GitFork, text: "parent:  bash → sshd (external SSH session)", color: COLORS.textDim, bold: false },
  { icon: Network, text: "network: → 45.33.32.156:3333 (mining pool)", color: COLORS.orange, bold: true },
  { icon: null, text: "", color: "", bold: false },
  { icon: AlertTriangle, text: "Verdict: crypto mining via unauthorized SSH", color: COLORS.accent, bold: true },
];

const TOOLS = ["summary", "inspect", "gonetwork_state", "family"];

export const LiveDemoScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const toolsDelay = 3 * fps;
  const thinkStart = 3.5 * fps;
  const thinkEnd = 4.2 * fps;
  const responseStart = 4.2 * fps;

  const headerOpacity = interpolate(frame, [0, 0.5 * fps], [0, 0.6], {
    extrapolateRight: "clamp",
    extrapolateLeft: "clamp",
  });

  return (
    <AbsoluteFill style={{ backgroundColor: COLORS.bg }}>
      {/* Top label */}
      <div
        style={{
          position: "absolute",
          top: 88,
          left: 0,
          right: 0,
          textAlign: "center",
          opacity: headerOpacity,
        }}
      >
        <span
          style={{
            fontFamily: FONT.sans,
            fontSize: 12,
            fontWeight: 300,
            color: COLORS.textDim,
            letterSpacing: 6,
            textTransform: "uppercase",
          }}
        >
          Live Investigation
        </span>
      </div>

      {/* Terminal */}
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Terminal title="godshell" width="78%">
          {/* Prompt */}
          <div style={{ marginBottom: 22 }}>
            <span style={{ color: COLORS.accent, fontWeight: 700, marginRight: 6 }}>❯</span>
            {frame < toolsDelay ? (
              <Typewriter
                text="why is my CPU usage so high?"
                charFrames={2}
                fontSize={22}
                color={COLORS.text}
                fontFamily={FONT.mono}
              />
            ) : (
              <span style={{ color: COLORS.text, fontSize: 22, fontFamily: FONT.mono }}>
                why is my CPU usage so high?
              </span>
            )}
          </div>

          {/* Tool badges */}
          {frame >= toolsDelay && (
            <div style={{ display: "flex", gap: 7, marginBottom: 20 }}>
              {TOOLS.map((tool, i) => {
                const badgeDelay = toolsDelay + i * 0.12 * fps;
                const badgeOpacity = interpolate(
                  frame - badgeDelay,
                  [0, 0.12 * fps],
                  [0, 1],
                  { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
                );
                return (
                  <span
                    key={tool}
                    style={{
                      opacity: badgeOpacity,
                      padding: "5px 12px",
                      borderRadius: 8,
                      background: COLORS.bgSubtle,
                      border: `1px solid ${COLORS.border}`,
                      fontFamily: FONT.mono,
                      fontSize: 12,
                      color: COLORS.textDim,
                      display: "inline-flex",
                      alignItems: "center",
                      gap: 5,
                    }}
                  >
                    <Activity size={10} color={COLORS.accent} strokeWidth={2} />
                    {tool}
                  </span>
                );
              })}
            </div>
          )}

          {/* Thinking dots */}
          {frame >= thinkStart && frame < thinkEnd && (
            <div style={{ display: "flex", gap: 6, marginBottom: 14 }}>
              {[0, 1, 2].map((dot) => {
                const dotOpacity = interpolate(
                  ((frame - thinkStart) / fps + dot * 0.2) % 0.6,
                  [0, 0.3, 0.6],
                  [0.3, 1, 0.3],
                  { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
                );
                return (
                  <div
                    key={dot}
                    style={{
                      width: 5,
                      height: 5,
                      borderRadius: "50%",
                      background: COLORS.accent,
                      opacity: dotOpacity,
                    }}
                  />
                );
              })}
            </div>
          )}

          {/* Response */}
          {frame >= responseStart && (
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: 4,
                borderLeft: `2px solid ${COLORS.border}`,
                paddingLeft: 18,
                marginLeft: 4,
              }}
            >
              {RESPONSE_LINES.map((line, i) => {
                const lineDelay = responseStart + i * 0.28 * fps;
                const lineOpacity = interpolate(
                  frame - lineDelay,
                  [0, 0.12 * fps],
                  [0, 1],
                  { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
                );
                const slideX = interpolate(
                  frame - lineDelay,
                  [0, 0.14 * fps],
                  [-8, 0],
                  { extrapolateRight: "clamp", extrapolateLeft: "clamp" }
                );

                if (!line.text) return <div key={i} style={{ height: 8 }} />;
                const LineIcon = line.icon;

                return (
                  <div
                    key={i}
                    style={{
                      opacity: lineOpacity,
                      transform: `translateX(${slideX}px)`,
                      fontFamily: FONT.mono,
                      fontSize: 17,
                      color: line.color,
                      fontWeight: line.bold ? 600 : 400,
                      lineHeight: 1.75,
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                    }}
                  >
                    {LineIcon && (
                      <LineIcon size={14} color={line.color} strokeWidth={1.5} style={{ flexShrink: 0 }} />
                    )}
                    <span>{line.text}</span>
                  </div>
                );
              })}
            </div>
          )}
        </Terminal>
      </div>

      {/* Letterbox */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
      <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: SPACING.letterbox, background: COLORS.bg }} />
    </AbsoluteFill>
  );
};
