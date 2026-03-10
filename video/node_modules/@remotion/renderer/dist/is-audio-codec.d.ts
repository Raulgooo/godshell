import type { Codec } from './codec';
export declare const isAudioCodec: (codec: Codec | undefined | null) => codec is "mp3" | "aac" | "wav";
