import type { Codec } from './codec';
import type { AudioCodec } from './options/audio-codec';
export declare const validateOutputFilename: <T extends Codec>({ codec, audioCodecSetting, extension, preferLossless, separateAudioTo, }: {
    codec: T;
    audioCodecSetting: AudioCodec | null;
    extension: string;
    preferLossless: boolean;
    separateAudioTo: string | null;
}) => void;
