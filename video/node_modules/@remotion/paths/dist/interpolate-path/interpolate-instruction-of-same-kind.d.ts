import type { CInstruction, LInstruction, MInstruction, ReducedInstruction, ZInstruction } from '../helpers/types';
export declare const interpolateInstructionOfSameKind: (t: number, first: ReducedInstruction, second: ReducedInstruction) => CInstruction | LInstruction | MInstruction | ZInstruction;
