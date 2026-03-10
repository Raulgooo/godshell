import type { ShapeInfo } from './shape-info';
type Direction = 'right' | 'left' | 'up' | 'down';
export type MakeTriangleProps = {
    length: number;
    direction: Direction;
    edgeRoundness?: number | null;
    cornerRadius?: number;
};
/**
 * @description Generates an SVG path for a triangle with the same length on all sides.
 * @param {Number} length The length of one triangle side.
 * @param {string} direction The direction of the triangle
 * @param {null|Number} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-triangle)
 */
export declare const makeTriangle: ({ length, direction, edgeRoundness, cornerRadius, }: MakeTriangleProps) => ShapeInfo;
export {};
