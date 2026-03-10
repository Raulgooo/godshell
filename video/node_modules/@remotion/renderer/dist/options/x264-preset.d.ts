import type { Codec } from '../codec';
export declare const x264PresetOptions: readonly ["ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow", "placebo"];
export type X264Preset = (typeof x264PresetOptions)[number];
export declare const validateSelectedCodecAndPresetCombination: ({ codec, x264Preset, }: {
    codec: Codec;
    x264Preset: X264Preset | null;
}) => void;
export declare const x264Option: {
    name: string;
    cliFlag: "x264-preset";
    description: () => import("react/jsx-runtime").JSX.Element;
    ssrName: "x264Preset";
    docLink: string;
    type: X264Preset | null;
    getValue: ({ commandLine }: {
        commandLine: Record<string, unknown>;
    }) => {
        value: X264Preset;
        source: string;
    } | {
        value: null;
        source: string;
    };
    setConfig: (profile: X264Preset | null) => void;
};
