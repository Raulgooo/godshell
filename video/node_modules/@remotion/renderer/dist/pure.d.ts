export declare const NoReactAPIs: {
    getExtensionOfFilename: (filename: string | null) => string | null;
    getFileExtensionFromCodec: <T extends import("./codec").Codec>(codec: T, audioCodec: import(".").AudioCodec | null) => import("./file-extensions").FileExtension;
    validateOutputFilename: <T extends import("./codec").Codec>({ codec, audioCodecSetting, extension, preferLossless, separateAudioTo, }: {
        codec: T;
        audioCodecSetting: import(".").AudioCodec | null;
        extension: string;
        preferLossless: boolean;
        separateAudioTo: string | null;
    }) => void;
    getFramesToRender: (frameRange: [number, number], everyNthFrame: number) => number[];
    codecSupportsMedia: (codec: import("./codec").Codec) => {
        video: boolean;
        audio: boolean;
    };
    isAudioCodec: (codec: import("./codec").Codec | undefined | null) => codec is "mp3" | "aac" | "wav";
};
