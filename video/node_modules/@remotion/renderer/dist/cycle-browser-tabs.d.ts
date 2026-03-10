import type { LogLevel } from './log-level';
import type { BrowserReplacer } from './replace-browser';
export declare const cycleBrowserTabs: ({ puppeteerInstance, concurrency, logLevel, indent, }: {
    puppeteerInstance: BrowserReplacer;
    concurrency: number;
    logLevel: LogLevel;
    indent: boolean;
}) => {
    stopCycling: () => void;
};
