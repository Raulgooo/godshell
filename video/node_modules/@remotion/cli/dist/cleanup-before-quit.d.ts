import type { LogLevel } from '@remotion/renderer';
export declare const cleanupBeforeQuit: ({ indent, logLevel, }: {
    indent: boolean;
    logLevel: LogLevel;
}) => void;
export declare const registerCleanupJob: (label: string, job: () => void) => void;
export declare const handleCtrlC: ({ indent, logLevel, }: {
    indent: boolean;
    logLevel: LogLevel;
}) => void;
