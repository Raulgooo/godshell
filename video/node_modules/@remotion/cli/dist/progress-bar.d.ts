import type { CancelSignal, LogLevel } from '@remotion/renderer';
import type { AggregateRenderProgress, BundlingState, CopyingState } from '@remotion/studio-server';
export type OverwriteableCliOutput = {
    update: (up: string, newline: boolean) => boolean;
};
export declare const LABEL_WIDTH = 20;
export declare const createOverwriteableCliOutput: (options: {
    quiet: boolean;
    cancelSignal: CancelSignal | null;
    updatesDontOverwrite: boolean;
    indent: boolean;
}) => OverwriteableCliOutput;
export type SymbolicLinksState = {
    symlinks: string[];
};
export declare const makeBundlingAndCopyProgress: ({ bundling, copying, symLinks, }: {
    bundling: BundlingState;
    copying: CopyingState;
    symLinks: SymbolicLinksState;
}) => string;
export declare const getRightLabelWidth: (totalFrames: number) => number;
export declare const makeRenderingAndStitchingProgress: ({ prog, isUsingParallelEncoding, }: {
    prog: AggregateRenderProgress;
    isUsingParallelEncoding: boolean;
}) => {
    output: string;
    progress: number;
    message: string;
};
export declare const printFact: (printLevel: LogLevel) => ({ indent, logLevel, left, right, color, link, }: {
    indent: boolean;
    logLevel: LogLevel;
    left: string;
    right: string;
    link?: string;
    color: "blue" | "blueBright" | "gray" | undefined;
}) => void;
