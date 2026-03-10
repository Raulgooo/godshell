import React from "react";
import { COLORS, FONT } from "../styles";

type TerminalProps = {
  children: React.ReactNode;
  title?: string;
  width?: number | string;
};

export const Terminal: React.FC<TerminalProps> = ({
  children,
  title = "godshell",
  width = "80%",
}) => {
  return (
    <div
      style={{
        width,
        backgroundColor: "#0A0A0A",
        borderRadius: 4,
        border: `1px solid #2A0000`,
        overflow: "hidden",
        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.8)",
      }}
    >
      {/* Title bar */}
      <div
        style={{
          backgroundColor: "#0F0000",
          height: 32,
          display: "flex",
          alignItems: "center",
          padding: "0 16px",
          borderBottom: `1px solid #1A0000`,
        }}
      >
        <div style={{ display: "flex", gap: 7 }}>
          <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: COLORS.textDim }} />
          <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: COLORS.textDim }} />
          <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: COLORS.textDim }} />
        </div>
        <div
          style={{
            flex: 1,
            textAlign: "center",
            color: COLORS.textSecondary,
            fontSize: 11,
            fontFamily: FONT.mono,
            opacity: 0.8,
            letterSpacing: 1
          }}
        >
          {title}
        </div>
        <div style={{ width: 44 }} />
      </div>

      {/* Content */}
      <div
        style={{
          padding: 24,
          fontFamily: FONT.mono,
          fontSize: 16,
          lineHeight: 1.7,
          background: "#0A0A0A",
          color: COLORS.text,
          minHeight: 200,
        }}
      >
        {children}
      </div>
    </div>
  );
};

