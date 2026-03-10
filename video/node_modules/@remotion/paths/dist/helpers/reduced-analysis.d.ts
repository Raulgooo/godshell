import type { Point, Properties, ReducedInstruction } from './types';
type SegmentInstruction = {
    function: Properties | null;
    length: number;
    instruction: ReducedInstruction;
    startPoint: Point;
};
type Segment = {
    startPoint: Point;
    instructionsAndInfo: SegmentInstruction[];
};
export declare const conductAnalysis: (instructions: ReducedInstruction[]) => Segment[];
export {};
