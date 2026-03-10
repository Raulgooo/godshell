import type { Constructed } from './helpers/construct';
type InstructionAtLength = {
    index: number;
    lengthIntoInstruction: number;
};
export declare const getInstructionIndexAtLengthFromConstructed: (constructed: Constructed, fractionLength: number) => InstructionAtLength;
export declare const getInstructionIndexAtLength: (path: string, length: number) => InstructionAtLength;
export {};
