import type { AbsoluteInstruction, QInstruction, ReducedInstruction } from './types';
export declare const iterateOverSegments: <T extends QInstruction | ReducedInstruction>({ segments, iterate, }: {
    segments: AbsoluteInstruction[];
    iterate: (options: {
        segment: AbsoluteInstruction;
        prevSegment: AbsoluteInstruction | null;
        x: number;
        y: number;
        initialX: number;
        initialY: number;
        cpX: number | null;
        cpY: number | null;
    }) => T[];
}) => T[];
