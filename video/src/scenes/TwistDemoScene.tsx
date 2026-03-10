import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";
import { Terminal } from "../components/Terminal";

export const TwistDemoScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const lines = [
    "> godshell config",
    "",
    "  [godshell] entering interactive config...",
    "  ? select default model: ",
    "  > gpt-4o",
    "    claude-3-5-sonnet",
    "    godshell-local-v1",
    "",
    "  [updating model to claude-3-5-sonnet...]",
    "",
    "  ⚠  ACCESS DENIED",
    "  Audit: Current process hierarchy is untrusted.",
    "  Cannot change system level config from this shell.",
  ];

  const linesShown = interpolate(frame, [1 * fps, 10 * fps], [0, lines.length], {
    extrapolateRight: "clamp",
  });

  const glitch = frame > 8 * fps && Math.random() > 0.9;

  return (
    <div style={{ flex: 1, backgroundColor: COLORS.bg, display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Terminal title="security_boundary.god" width={900}>
        <div style={{ color: COLORS.text, filter: glitch ? "invert(1) hue-rotate(90deg)" : "none" }}>
          {lines.slice(0, Math.floor(linesShown)).map((line, i) => {
             const isDenied = line.includes("DENIED");
             const color = isDenied ? COLORS.red : COLORS.text;

             return (
               <div key={i} style={{ 
                 color, 
                 fontWeight: isDenied ? 800 : 400,
                 marginBottom: line === "" ? 12 : 4,
                 opacity: interpolate(linesShown - i, [0, 0.5], [0, 1])
               }}>
                 {line}
               </div>
             );
          })}
        </div>
      </Terminal>

      {frame > 10 * fps && (
        <div style={{ 
          position: "absolute", 
          backgroundColor: COLORS.red, 
          color: "white", 
          padding: "20px 40px",
          fontFamily: FONT.sans,
          fontSize: 32,
          fontWeight: 800,
          transform: `rotate(-5deg) scale(${spring({ frame: frame - 10 * fps, fps })})`,
          boxShadow: `0 0 50px ${COLORS.red}`
        }}>
          TRUST NO ONE
        </div>
      )}
    </div>
  );
};
