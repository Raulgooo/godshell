import type { ShapeInfo } from './shape-info';
export type MakeStarProps = {
    points: number;
    innerRadius: number;
    outerRadius: number;
    edgeRoundness?: number | null;
    cornerRadius?: number;
};
export type StarProps = {
    centerX: number;
    centerY: number;
    points: number;
    innerRadius: number;
    outerRadius: number;
    edgeRoundness: number | null;
    cornerRadius: number;
};
export declare const makeStar: ({ points, innerRadius, outerRadius, cornerRadius, edgeRoundness, }: MakeStarProps) => ShapeInfo;
