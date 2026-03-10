import type { LogLevel } from '@remotion/renderer';
import type { GitSource } from '@remotion/studio-shared';
export type ParsedGitRemote = {
    type: 'github';
    org: string;
    name: string;
};
export declare const getGitConfig: (remotionRoot: string) => string | null;
export declare const getGitRemoteOrigin: (gitConfig: string) => {
    remote: string;
    url: string | null;
} | null;
export declare const normalizeGitRemoteUrl: (url: string) => ParsedGitRemote | null;
export declare const getGifRef: (logLevel: LogLevel) => string | null;
export declare const getGitSource: ({ remotionRoot, disableGitSource, logLevel, }: {
    remotionRoot: string;
    disableGitSource: boolean;
    logLevel: LogLevel;
}) => GitSource | null;
