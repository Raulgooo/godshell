import type { Instruction } from '@remotion/paths';
export declare const joinPoints: (points: [number, number][], { edgeRoundness, cornerRadius, roundCornerStrategy, }: {
    edgeRoundness: number | null;
    cornerRadius: number;
    roundCornerStrategy: "arc" | "bezier";
}) => Instruction[];
