import type { GetSilentPartsResponse, SilentPart } from './compositor/payloads';
import type { LogLevel } from './log-level';
export type { SilentPart };
export declare const getSilentParts: ({ src, noiseThresholdInDecibels: passedNoiseThresholdInDecibels, minDurationInSeconds: passedMinDuration, logLevel, binariesDirectory, }: {
    src: string;
    minDurationInSeconds?: number;
    logLevel?: LogLevel;
    noiseThresholdInDecibels?: number;
    binariesDirectory?: string | null;
}) => Promise<GetSilentPartsResponse>;
