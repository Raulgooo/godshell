import type { BrowserExecutable } from './browser-executable';
import type { LogLevel } from './log-level';
import type { ChromeMode } from './options/chrome-mode';
export declare const getLocalBrowserExecutable: ({ preferredBrowserExecutable, logLevel, indent, chromeMode, }: {
    preferredBrowserExecutable: BrowserExecutable;
    logLevel: LogLevel;
    indent: boolean;
    chromeMode: ChromeMode;
}) => string;
