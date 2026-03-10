import type { ShapeInfo } from './shape-info';
export type MakeRectOptions = {
    width: number;
    height: number;
    edgeRoundness?: number | null;
    cornerRadius?: number;
};
/**
 * @description Generates an SVG rectangle.
 * @param {Number} width The width of the rectangle.
 * @param {Number} height The height of the rectangle
 * @param {Number | null} edgeRoundness Allows to modify the shape by rounding the edges using bezier curves. Default null.
 * @param {Number} cornerRadius  * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius. Cannot be used together with edgeRoundness.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-rect)
 */
export declare const makeRect: ({ width, height, edgeRoundness, cornerRadius, }: MakeRectOptions) => ShapeInfo;
