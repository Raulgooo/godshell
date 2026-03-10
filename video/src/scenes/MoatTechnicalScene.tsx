import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";

export const MoatTechnicalScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const nodes = [
    { label: "USER SPACE", x: -200, y: -100, color: COLORS.textSecondary },
    { label: "GODSHELL CORE", x: 0, y: 0, color: COLORS.accent },
    { label: "KERNEL SPACE", x: 200, y: 100, color: COLORS.red },
    { label: "eBPF SENSORS", x: 200, y: -100, color: COLORS.accentMuted },
  ];

  const progress = spring({
    frame,
    fps,
    config: { damping: 100 },
  });

  return (
    <div style={{ flex: 1, backgroundColor: COLORS.bgDeep, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center" }}>
      <h2 style={{ 
        fontFamily: FONT.sans, 
        fontSize: 48, 
        fontWeight: 200, 
        color: COLORS.text, 
        marginBottom: 80,
        opacity: interpolate(frame, [0, 1 * fps], [0, 1]),
        letterSpacing: 4
      }}>
        THE MOAT
      </h2>

      <div style={{ position: "relative", width: 800, height: 400 }}>
        {/* Draw connectors */}
        <svg style={{ position: "absolute", top: 0, left: 0, width: "100%", height: "100%", overflow: "visible" }}>
           <line x1="400" y1="200" x2="200" y2="100" stroke={COLORS.accentSubtle} strokeWidth={2} opacity={progress} />
           <line x1="400" y1="200" x2="600" y2="300" stroke={COLORS.accentSubtle} strokeWidth={2} opacity={progress} />
           <line x1="400" y1="200" x2="600" y2="100" stroke={COLORS.accentSubtle} strokeWidth={2} opacity={progress} />
        </svg>

        {nodes.map((node, i) => {
          const nodeSpring = spring({
            frame: frame - i * 10,
            fps,
          });
          
          return (
            <div key={i} style={{
              position: "absolute",
              left: 400 + node.x - 100,
              top: 200 + node.y - 40,
              width: 200,
              height: 80,
              backgroundColor: COLORS.bgSubtle,
              border: `1px solid ${node.color}`,
              borderRadius: 4,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: node.color,
              fontFamily: FONT.mono,
              fontSize: 14,
              fontWeight: 800,
              transform: `scale(${nodeSpring})`,
              boxShadow: `0 0 20px ${node.color}33`
            }}>
              {node.label}
            </div>
          );
        })}
      </div>

      <div style={{ 
        marginTop: 100, 
        fontFamily: FONT.mono, 
        color: COLORS.textSecondary,
        fontSize: 18,
        opacity: interpolate(frame, [4 * fps, 6 * fps], [0, 1])
      }}>
        Bypassing standard audit logs. Direct kernel visibility.
      </div>
    </div>
  );
};
