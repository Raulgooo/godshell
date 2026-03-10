import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";
import { Terminal } from "../components/Terminal";

export const MemoryHeroDemoScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const lines = [
    "> suspicious binary in /tmp, pid 302286",
    "",
    "  [tool: inspect 302286]",
    "  [tool: read_memory 302286 --offset 0x4000]",
    "",
    "  Reading memory pages...",
    "  0x4000: 47 4F 44 53 48 45 4C 4C  GODSHELL",
    "  0x4008: 5F 4C 49 43 45 4E 53 45  _LICENSE",
    "  0x4010: 5F 4B 45 59 3D 61 62 63  _KEY=abc",
    "  0x4018: 31 32 33 2D 78 79 7A 37  123-xyz7",
    "",
    "⚠  SENSITIVE DATA LEAK DETECTED",
    "License key found in plaintext in /tmp/rev_shell_v2",
  ];

  const linesShown = interpolate(frame, [1 * fps, 15 * fps], [0, lines.length], {
    extrapolateRight: "clamp",
  });

  return (
    <div style={{ flex: 1, backgroundColor: COLORS.bg, display: "flex", alignItems: "center", justifyContent: "center" }}>
      <Terminal title="memory_audit.god" width={1100}>
        <div style={{ color: COLORS.text }}>
          {lines.slice(0, Math.floor(linesShown)).map((line, i) => {
             const isWarning = line.includes("DETECTED");
             const isHexLine = line.includes("0x40");
             const isCommand = line.startsWith(">");
             
             let color = COLORS.text;
             if (isWarning) color = COLORS.red;
             if (isHexLine) color = COLORS.textSecondary;
             if (isCommand) color = COLORS.accent;

             return (
               <div key={i} style={{ 
                 color, 
                 fontWeight: isWarning ? 800 : 400,
                 fontFamily: isHexLine ? FONT.mono : FONT.mono,
                 marginBottom: line === "" ? 12 : 2,
                 opacity: interpolate(linesShown - i, [0, 0.5], [0, 1])
               }}>
                 {line}
               </div>
             );
          })}
        </div>
      </Terminal>
    </div>
  );
};
