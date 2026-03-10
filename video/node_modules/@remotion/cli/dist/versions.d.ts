import type { LogLevel } from '@remotion/renderer';
export type VersionAndPath = {
    version: string;
    path: string;
};
export declare const VERSIONS_COMMAND = "versions";
export declare const validateVersionsBeforeCommand: (remotionRoot: string, logLevel: LogLevel) => Promise<void>;
export declare const versionsCommand: (remotionRoot: string, logLevel: LogLevel) => Promise<void>;
