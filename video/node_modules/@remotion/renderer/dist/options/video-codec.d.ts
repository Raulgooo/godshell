import type { Codec, CodecOrUndefined } from '../codec';
export declare const getOutputCodecOrUndefined: () => CodecOrUndefined;
export declare const videoCodecOption: {
    name: string;
    cliFlag: "codec";
    description: () => import("react/jsx-runtime").JSX.Element;
    ssrName: string;
    docLink: string;
    type: Codec;
    getValue: ({ commandLine }: {
        commandLine: Record<string, unknown>;
    }, { compositionCodec, configFile, downloadName, outName, uiCodec, }: {
        outName: string | null;
        downloadName: string | null;
        configFile: Codec | null;
        uiCodec: Codec | null;
        compositionCodec: Codec | null;
    }) => {
        value: Codec;
        source: string;
    };
    setConfig: (newCodec: CodecOrUndefined) => void;
};
