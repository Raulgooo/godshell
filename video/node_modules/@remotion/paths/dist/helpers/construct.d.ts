import type { Instruction, Point, Properties } from './types';
export type Constructed = {
    segments: Instruction[][];
    initialPoint: Point | null;
    totalLength: number;
    partialLengths: number[];
    functions: (Properties | null)[];
};
export declare const constructFromInstructions: (instructions: Instruction[]) => Constructed;
export declare const construct: (string: string) => Constructed;
