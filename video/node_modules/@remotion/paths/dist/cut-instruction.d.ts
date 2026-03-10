import type { CInstruction, Point, ReducedInstruction } from './helpers/types';
export declare function cutCInstruction({ progress, lastPoint, instruction }: {
    progress: number;
    lastPoint: Point;
    instruction: CInstruction;
}): CInstruction;
export declare const cutInstruction: ({ instruction, lastPoint, progress, }: {
    instruction: ReducedInstruction;
    lastPoint: Point;
    progress: number;
}) => ReducedInstruction;
