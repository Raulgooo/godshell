import React from "react";
import {
  useCurrentFrame,
  useVideoConfig,
  interpolate,
} from "remotion";
import { COLORS, FONT } from "../styles";

type TypewriterProps = {
  text: string;
  pauseAfter?: string;
  charFrames?: number;
  pauseSeconds?: number;
  fontSize?: number;
  color?: string;
  fontFamily?: string;
  cursorSymbol?: string;
};

const getTypedText = ({
  frame,
  fullText,
  pauseAfter,
  charFrames,
  pauseFrames,
}: {
  frame: number;
  fullText: string;
  pauseAfter: string;
  charFrames: number;
  pauseFrames: number;
}): string => {
  const pauseIndex = fullText.indexOf(pauseAfter);
  const preLen =
    pauseIndex >= 0 ? pauseIndex + pauseAfter.length : fullText.length;

  let typedChars = 0;
  if (frame < preLen * charFrames) {
    typedChars = Math.floor(frame / charFrames);
  } else if (frame < preLen * charFrames + pauseFrames) {
    typedChars = preLen;
  } else {
    const postPhase = frame - preLen * charFrames - pauseFrames;
    typedChars = Math.min(
      fullText.length,
      preLen + Math.floor(postPhase / charFrames)
    );
  }
  return fullText.slice(0, typedChars);
};

export const Typewriter: React.FC<TypewriterProps> = ({
  text,
  pauseAfter = "",
  charFrames = 2,
  pauseSeconds = 0,
  fontSize = 64,
  color = COLORS.text,
  fontFamily = FONT.mono,
  cursorSymbol = "▌",
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const pauseFrames = Math.round(fps * pauseSeconds);

  const typedText = getTypedText({
    frame,
    fullText: text,
    pauseAfter: pauseAfter || text,
    charFrames,
    pauseFrames,
  });

  const cursorOpacity = interpolate(
    frame % 16,
    [0, 8, 16],
    [1, 0, 1],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );

  return (
    <div
      style={{
        fontSize,
        fontFamily,
        color,
        fontWeight: 600,
        lineHeight: 1.4,
        whiteSpace: "pre-wrap",
      }}
    >
      <span>{typedText}</span>
      <span style={{ opacity: cursorOpacity }}>{cursorSymbol}</span>
    </div>
  );
};
