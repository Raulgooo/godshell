import type { LogLevel } from '@remotion/renderer';
type FoundReason = 'argument passed - found in cwd' | 'argument passed - found in root' | 'argument passed' | 'config file' | 'common paths' | 'none found';
export declare const findEntryPoint: ({ args, logLevel, remotionRoot, allowDirectory, }: {
    args: (string | number)[];
    remotionRoot: string;
    logLevel: LogLevel;
    allowDirectory: boolean;
}) => {
    file: string | null;
    remainingArgs: (string | number)[];
    reason: FoundReason;
};
export {};
