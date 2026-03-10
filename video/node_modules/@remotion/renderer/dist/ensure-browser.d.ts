import type { BrowserExecutable } from './browser-executable';
import type { BrowserSafeApis } from './client';
import type { ChromeMode } from './options/chrome-mode';
import type { ToOptions } from './options/option';
export type BrowserStatus = {
    type: 'user-defined-path';
    path: string;
} | {
    type: 'local-puppeteer-browser';
    path: string;
} | {
    type: 'no-browser';
};
type InternalEnsureBrowserOptions = {
    browserExecutable: BrowserExecutable;
    indent: boolean;
    chromeMode: ChromeMode;
} & ToOptions<typeof BrowserSafeApis.optionsMap.ensureBrowser>;
export type EnsureBrowserOptions = Partial<{
    browserExecutable: BrowserExecutable;
} & ToOptions<typeof BrowserSafeApis.optionsMap.ensureBrowser>>;
export declare const internalEnsureBrowser: (options: InternalEnsureBrowserOptions) => Promise<BrowserStatus>;
export declare const ensureBrowser: (options?: EnsureBrowserOptions) => Promise<BrowserStatus>;
export {};
