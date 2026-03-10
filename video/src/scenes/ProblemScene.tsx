import React from "react";
import { useCurrentFrame, useVideoConfig, interpolate, spring } from "remotion";
import { COLORS, FONT } from "../styles";
import { Terminal } from "../components/Terminal";

export const ProblemScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // ps aux scroll effect
  const scrollOffset = interpolate(frame, [0, 8 * fps], [0, -1000], {
    extrapolateRight: "clamp",
  });

  // Typewriter text
  const text1 = "Your system is talking.";
  const text2 = "You just can't hear it.";
  
  const charsShown1 = interpolate(frame, [1 * fps, 3 * fps], [0, text1.length], {
    extrapolateRight: "clamp",
  });
  const charsShown2 = interpolate(frame, [3.5 * fps, 5.5 * fps], [0, text2.length], {
    extrapolateRight: "clamp",
  });

  const overlayOpacity = interpolate(frame, [0, 1 * fps], [0, 1]);

  const psLines = [
    "USER   PID  %CPU  COMMAND",
    "raul   1234  0.0  /usr/lib/systemd/systemd --user",
    "raul   5678  2.1  /opt/google/chrome/chrome",
    "raul   9012  0.3  python3 crypto_miner.py",
    "root    234  0.0  [kworker/0:1]",
    "root    891  0.0  [kworker/1:2H]",
    "raul  11203  0.1  /usr/bin/gnome-shell",
    "raul  15423  0.0  sshd: raul@pts/0",
    "root   1232  0.0  /usr/bin/containerd",
    "raul  23124  4.5  node /path/to/server.js",
    "raul  12312  1.2  /usr/bin/python3 suspicious.py",
    "raul   4321  0.1  /usr/bin/bash",
    "root    442  0.0  [jbd2/nvme0n1p2-]",
    "raul   8823  0.0  /usr/lib/gdm3/gdm-wayland-session",
    "raul   9921  3.4  docker-containerd",
  ];

  // Repeat lines to simulate growth
  const allLines = Array(20).fill(psLines).flat();

  return (
    <div style={{ flex: 1, backgroundColor: COLORS.bg, display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center" }}>
      <div style={{ opacity: 0.4, transform: `translateY(${scrollOffset}px)`, transition: "none" }}>
        <Terminal title="ps aux" width={1000}>
          <div style={{ fontSize: 14, color: COLORS.textSecondary }}>
            {allLines.map((line, i) => (
              <div key={i}>{line}</div>
            ))}
          </div>
        </Terminal>
      </div>

      <div style={{ 
        position: "absolute", 
        textAlign: "center",
        opacity: overlayOpacity,
        textShadow: "0 0 40px rgba(8, 8, 8, 1), 0 0 100px rgba(8, 8, 8, 1)"
      }}>
        <h2 style={{ 
          fontFamily: FONT.sans, 
          fontSize: 64, 
          fontWeight: 200, 
          color: COLORS.text, 
          margin: 0,
          letterSpacing: -1
        }}>
          {text1.slice(0, Math.floor(charsShown1))}
        </h2>
        <h2 style={{ 
          fontFamily: FONT.sans, 
          fontSize: 64, 
          fontWeight: 200, 
          color: COLORS.accent, 
          margin: 0,
          marginTop: 10,
          letterSpacing: -1
        }}>
          {text2.slice(0, Math.floor(charsShown2))}
        </h2>
        {frame > 5.5 * fps && (
           <div style={{ 
             width: 4, height: 60, backgroundColor: COLORS.accent, display: "inline-block", 
             verticalAlign: "middle", marginLeft: 8,
             opacity: Math.floor((frame / fps) * 2) % 2 === 0 ? 1 : 0
           }} />
        )}
      </div>
    </div>
  );
};
