import type { LogLevel } from './log-level';
import type { CancelSignal } from './make-cancel-signal';
export declare const muxVideoAndAudio: ({ videoOutput, audioOutput, output, indent, logLevel, onProgress, binariesDirectory, fps, cancelSignal, addFaststart, metadata, }: {
    videoOutput: string | null;
    audioOutput: string | null;
    output: string;
    indent: boolean;
    logLevel: LogLevel;
    binariesDirectory: string | null;
    fps: number;
    onProgress: (p: number) => void;
    cancelSignal: CancelSignal | undefined;
    addFaststart: boolean;
    metadata?: Record<string, string> | null;
}) => Promise<void>;
