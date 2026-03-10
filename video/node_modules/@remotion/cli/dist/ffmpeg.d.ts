import type { LogLevel } from '@remotion/renderer';
export declare const dynamicLibEnv: (indent: boolean, logLevel: LogLevel, binariesDirectory: string | null) => {
    DYLD_LIBRARY_PATH: string;
    RUST_BACKTRACE: string;
} | {
    PATH: string;
    RUST_BACKTRACE: string;
} | {
    LD_LIBRARY_PATH: string;
    RUST_BACKTRACE: string;
};
export declare const ffmpegCommand: (args: string[], logLevel: LogLevel) => never;
export declare const ffprobeCommand: (args: string[], logLevel: LogLevel) => never;
