import React from "react";
import { TransitionSeries, springTiming } from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { slide } from "@remotion/transitions/slide";
import { useVideoConfig } from "remotion";

import { ColdOpenScene } from "./scenes/ColdOpenScene";
import { ProblemScene } from "./scenes/ProblemScene";
import { BrandScene } from "./scenes/BrandScene";
import { ExfiltrationDemoScene } from "./scenes/ExfiltrationDemoScene";
import { MemoryHeroDemoScene } from "./scenes/MemoryHeroDemoScene";
import { TwistDemoScene } from "./scenes/TwistDemoScene";
import { MoatTechnicalScene } from "./scenes/MoatTechnicalScene";
import { FinalCTAScene } from "./scenes/FinalCTAScene";
import { DURATIONS } from "./styles";

export const GodshellTrailer: React.FC = () => {
  const { fps } = useVideoConfig();

  const sec = (s: number) => Math.round(s * fps);

  // Cinematic transitions
  const springT = springTiming({
    config: { damping: 200 },
    durationInFrames: 15,
  });

  const slowFade = springTiming({
    config: { damping: 200 },
    durationInFrames: 30,
  });

  return (
    <TransitionSeries>
      {/* Scene 1: Cold Open */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.coldOpen)}>
        <ColdOpenScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={slide({ direction: "from-bottom" })}
        timing={springT}
      />

      {/* Scene 2: Problem */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.problem)}>
        <ProblemScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={fade()}
        timing={springT}
      />

      {/* Scene 3: Brand Reveal */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.brand)}>
        <BrandScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={slide({ direction: "from-right" })}
        timing={springT}
      />

      {/* Scene 4: Demo 1 — Exfiltration */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.demo1)}>
        <ExfiltrationDemoScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={slide({ direction: "from-right" })}
        timing={springT}
      />

      {/* Scene 5: Demo 2 — Memory */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.demo2)}>
        <MemoryHeroDemoScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={slide({ direction: "from-right" })}
        timing={springT}
      />

      {/* Scene 6: Demo 3 — Twist */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.demo3)}>
        <TwistDemoScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={fade()}
        timing={slowFade}
      />

      {/* Scene 7: Moat Technical */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.moat)}>
        <MoatTechnicalScene />
      </TransitionSeries.Sequence>

      <TransitionSeries.Transition
        presentation={fade()}
        timing={springT}
      />

      {/* Scene 8: Final CTA */}
      <TransitionSeries.Sequence durationInFrames={sec(DURATIONS.cta)}>
        <FinalCTAScene />
      </TransitionSeries.Sequence>
    </TransitionSeries>
  );
};

