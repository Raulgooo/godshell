import type { ShapeInfo } from './shape-info';
export type MakeCircleProps = {
    radius: number;
};
/**
 * @description Generates a circle SVG path.
 * @param {Number} radius The radius of the circle.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-circle)
 */
export declare const makeCircle: ({ radius }: MakeCircleProps) => ShapeInfo;
