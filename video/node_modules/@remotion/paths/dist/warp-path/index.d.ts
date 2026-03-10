import type { WarpPathFn } from './warp-helpers';
export declare const warpPath: (path: string, transformer: WarpPathFn, options?: {
    interpolationThreshold?: number | undefined;
} | undefined) => string;
export type { WarpPathFn } from './warp-helpers';
