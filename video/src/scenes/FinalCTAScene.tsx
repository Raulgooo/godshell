import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";

export const FinalCTAScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoSpring = spring({
    frame,
    fps,
    config: { damping: 100, stiffness: 40 },
  });

  const opacity = interpolate(frame, [0, 1 * fps], [0, 1]);
  const scale = interpolate(logoSpring, [0, 1], [0.8, 1]);

  return (
    <div
      style={{
        flex: 1,
        backgroundColor: COLORS.bgDeep,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div style={{ opacity, transform: `scale(${scale})`, textAlign: "center" }}>
        <h1
          style={{
            fontFamily: FONT.sans,
            fontSize: 160,
            fontWeight: 800,
            color: COLORS.red,
            letterSpacing: 20,
            margin: 0,
            filter: `drop-shadow(0 0 50px rgba(255, 34, 34, 0.6))`,
          }}
        >
          GODSHELL
        </h1>
        
        <p
          style={{
            fontFamily: FONT.mono,
            fontSize: 32,
            color: COLORS.textSecondary,
            letterSpacing: 2,
            marginTop: 40,
            fontWeight: 200
          }}
        >
          One command for total system awareness.
        </p>

        <div style={{ marginTop: 80, display: "flex", gap: 20, justifyContent: "center" }}>
           <div style={{ 
             backgroundColor: COLORS.red, 
             color: "white", 
             padding: "20px 40px", 
             borderRadius: 100, 
             fontFamily: FONT.mono, 
             fontSize: 24, 
             fontWeight: 800,
             boxShadow: `0 0 30px ${COLORS.red}66`
           }}>
             godshell.com
           </div>
        </div>

        <div style={{ 
          marginTop: 60, 
          color: COLORS.textDim, 
          fontFamily: FONT.mono, 
          fontSize: 14,
          opacity: 0.6
        }}>
          # TRUST BUT VERIFY
        </div>
      </div>
    </div>
  );
};
