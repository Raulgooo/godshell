import { Composition } from "remotion";
import { GodshellTrailer } from "./GodshellTrailer";
import { VIDEO, DURATIONS } from "./styles";

// Calculate total duration accounting for transitions
// 8 scenes with 7 transitions of 15 frames (~0.5s) each
const totalSeconds =
  DURATIONS.coldOpen +
  DURATIONS.problem +
  DURATIONS.brand +
  DURATIONS.demo1 +
  DURATIONS.demo2 +
  DURATIONS.demo3 +
  DURATIONS.moat +
  DURATIONS.cta;

const transitionOverlap = 7 * 0.5; // 7 transitions × 0.5s each
const finalDuration = totalSeconds - transitionOverlap;
const totalFrames = Math.round(finalDuration * VIDEO.fps);

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="GodshellTrailer"
      component={GodshellTrailer}
      durationInFrames={totalFrames}
      fps={VIDEO.fps}
      width={VIDEO.width}
      height={VIDEO.height}
    />
  );
};

