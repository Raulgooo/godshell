import type { ReducedInstruction } from '../helpers/types';
export type WarpPathFn = (point: {
    x: number;
    y: number;
}) => {
    x: number;
    y: number;
};
export declare function svgPathInterpolate(path: ReducedInstruction[], threshold: number): ReducedInstruction[];
export declare const warpTransform: (path: ReducedInstruction[], transformer: WarpPathFn) => ReducedInstruction[];
export declare const fixZInstruction: (instructions: ReducedInstruction[]) => ReducedInstruction[];
