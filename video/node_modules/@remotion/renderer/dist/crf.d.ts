import type { HardwareAccelerationOption } from './client';
import type { Codec } from './codec';
export type Crf = number | undefined;
export declare const getDefaultCrfForCodec: (codec: Codec) => number | null;
export declare const getValidCrfRanges: (codec: Codec) => [number, number];
export declare const validateQualitySettings: ({ codec, crf, videoBitrate, encodingMaxRate, encodingBufferSize, hardwareAcceleration, }: {
    crf: unknown;
    codec: Codec;
    videoBitrate: string | null;
    encodingMaxRate: string | null;
    encodingBufferSize: string | null;
    hardwareAcceleration: HardwareAccelerationOption;
}) => string[];
