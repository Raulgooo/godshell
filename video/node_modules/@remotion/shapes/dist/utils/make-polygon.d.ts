import type { ShapeInfo } from './shape-info';
export type MakePolygonProps = {
    points: number;
    radius: number;
    edgeRoundness?: number | null;
    cornerRadius?: number;
};
export type PolygonProps = {
    centerX: number;
    centerY: number;
    points: number;
    radius: number;
    edgeRoundness: number | null;
    cornerRadius: number;
};
export declare const makePolygon: ({ points, radius, cornerRadius, edgeRoundness, }: MakePolygonProps) => ShapeInfo;
