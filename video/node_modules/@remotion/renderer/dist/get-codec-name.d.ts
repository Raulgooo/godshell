import type { Codec } from './codec';
import type { LogLevel } from './log-level';
import type { HardwareAccelerationOption } from './options/hardware-acceleration';
export type CodecSettings = {
    encoderName: string;
    hardwareAccelerated: boolean;
};
export declare const hasSpecifiedUnsupportedHardwareQualifySettings: ({ encodingMaxRate, encodingBufferSize, crf, }: {
    encodingMaxRate: string | null;
    encodingBufferSize: string | null;
    crf: unknown;
}) => "crf" | "encodingBufferSize" | "encodingMaxRate" | null;
export declare const getCodecName: ({ codec, encodingMaxRate, encodingBufferSize, crf, hardwareAcceleration, logLevel, indent, }: {
    codec: Codec;
    hardwareAcceleration: HardwareAccelerationOption;
    encodingMaxRate: string | null;
    encodingBufferSize: string | null;
    crf: unknown;
    logLevel: LogLevel;
    indent: boolean;
}) => CodecSettings | null;
