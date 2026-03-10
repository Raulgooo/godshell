import { type LogLevel } from '@remotion/renderer';
export declare const upgradeCommand: ({ remotionRoot, packageManager, version, logLevel, args, }: {
    remotionRoot: string;
    packageManager: string | undefined;
    version: string | undefined;
    logLevel: LogLevel;
    args: string[];
}) => Promise<void>;
