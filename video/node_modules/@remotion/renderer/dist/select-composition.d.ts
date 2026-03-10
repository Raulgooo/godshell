import type { VideoConfig } from 'remotion/no-react';
import type { BrowserExecutable } from './browser-executable';
import type { BrowserLog } from './browser-log';
import type { HeadlessBrowser } from './browser/Browser';
import type { ChromiumOptions } from './open-browser';
import type { ToOptions } from './options/option';
import type { optionsMap } from './options/options-map';
import type { RemotionServer } from './prepare-server';
import type { RequiredInputPropsInV5 } from './v5-required-input-props';
type InternalSelectCompositionsConfig = {
    serializedInputPropsWithCustomSchema: string;
    envVariables: Record<string, string>;
    puppeteerInstance: HeadlessBrowser | undefined;
    onBrowserLog: null | ((log: BrowserLog) => void);
    browserExecutable: BrowserExecutable | null;
    chromiumOptions: ChromiumOptions;
    port: number | null;
    indent: boolean;
    server: RemotionServer | undefined;
    serveUrl: string;
    id: string;
    onServeUrlVisited: () => void;
} & ToOptions<typeof optionsMap.selectComposition>;
export type SelectCompositionOptions = RequiredInputPropsInV5 & {
    envVariables?: Record<string, string>;
    puppeteerInstance?: HeadlessBrowser;
    onBrowserLog?: (log: BrowserLog) => void;
    browserExecutable?: BrowserExecutable;
    chromiumOptions?: ChromiumOptions;
    port?: number | null;
    /**
     * @deprecated Use `logLevel` instead.
     */
    verbose?: boolean;
    serveUrl: string;
    id: string;
} & Partial<ToOptions<typeof optionsMap.selectComposition>>;
type InternalReturnType = {
    metadata: VideoConfig;
    propsSize: number;
};
export declare const internalSelectCompositionRaw: (options: InternalSelectCompositionsConfig) => Promise<InternalReturnType>;
export declare const internalSelectComposition: (options: InternalSelectCompositionsConfig) => Promise<InternalReturnType>;
export declare const selectComposition: (options: SelectCompositionOptions) => Promise<VideoConfig>;
export {};
