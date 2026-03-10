import { NoReactInternals } from 'remotion/no-react';
declare const validV4ColorSpaces: readonly ["default", "bt709", "bt2020-ncl"];
declare const validV5ColorSpaces: readonly ["bt601", "bt709", "bt2020-ncl"];
export declare const validColorSpaces: true extends typeof NoReactInternals.ENABLE_V5_BREAKING_CHANGES ? typeof validV5ColorSpaces : typeof validV4ColorSpaces;
export type ColorSpace = (typeof validColorSpaces)[number];
export declare const DEFAULT_COLOR_SPACE: true extends typeof NoReactInternals.ENABLE_V5_BREAKING_CHANGES ? "bt709" : "default";
export declare const colorSpaceOption: {
    name: string;
    cliFlag: "color-space";
    description: () => import("react/jsx-runtime").JSX.Element;
    docLink: string;
    ssrName: string;
    type: ColorSpace | null;
    getValue: ({ commandLine }: {
        commandLine: Record<string, unknown>;
    }) => {
        source: string;
        value: ColorSpace;
    };
    setConfig: (value: "default" | "bt709" | "bt2020-ncl" | null) => void;
};
export declare const validateColorSpace: (option: unknown) => void;
export {};
