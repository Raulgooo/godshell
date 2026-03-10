import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";
import { Terminal } from "../components/Terminal";

export const ExfiltrationDemoScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const lines = [
    "> hey, anything weird on my system?",
    "",
    "  [tool: summary]",
    "  [tool: inspect 137624]",
    "",
    "⚠  HIGH SEVERITY",
    "",
    "python3 (137624) read ~/.ssh/id_rsa",
    "then connected to:",
    "ec2-18-97-36-17.compute-1.amazonaws.com:443",
    "",
    "Recommendation: SSH key exfiltration.",
    "Kill PID 137624 immediately.",
  ];

  const linesShown = interpolate(frame, [1 * fps, 10 * fps], [0, lines.length], {
    extrapolateRight: "clamp",
  });

  const severityPulse = spring({
    frame: frame - 6 * fps,
    fps,
    config: { damping: 10, stiffness: 40 },
  });

  const flashOpacity = interpolate(severityPulse, [0, 1], [0, 0.4]);

  return (
    <div style={{ flex: 1, backgroundColor: COLORS.bg, display: "flex", alignItems: "center", justifyContent: "center" }}>
      {/* Background flash for high severity */}
      <div style={{ 
        position: "absolute", 
        inset: 0, 
        backgroundColor: COLORS.red, 
        opacity: frame > 6 * fps ? Math.sin(frame / 5) * 0.1 : 0 
      }} />

      <Terminal title="exfiltration_check.god" width={1000}>
        <div style={{ color: COLORS.text }}>
          {lines.slice(0, Math.floor(linesShown)).map((line, i) => {
             const isSeverity = line.includes("HIGH SEVERITY");
             const isRecommendation = line.includes("Recommendation") || line.includes("Kill");
             const color = isSeverity ? COLORS.red : isRecommendation ? COLORS.accent : COLORS.text;
             const weight = isSeverity || isRecommendation ? 800 : 400;

             return (
               <div key={i} style={{ 
                 color, 
                 fontWeight: weight, 
                 marginBottom: line === "" ? 12 : 4,
                 textShadow: isSeverity ? `0 0 20px ${COLORS.red}` : "none",
                 opacity: interpolate(linesShown - i, [0, 0.5], [0, 1])
               }}>
                 {line}
               </div>
             );
          })}
          {linesShown < lines.length && (
             <span style={{ 
               display: "inline-block", backgroundColor: COLORS.accent, width: 8, height: 18, verticalAlign: "middle", marginLeft: 4,
               opacity: Math.floor((frame / fps) * 2) % 2 === 0 ? 1 : 0
             }} />
          )}
        </div>
      </Terminal>
    </div>
  );
};
