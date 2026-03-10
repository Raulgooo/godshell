import type { ShapeInfo } from './shape-info';
export type MakeEllipseOptions = {
    rx: number;
    ry: number;
};
/**
 * @description Generates an ellipse SVG path.
 * @param {Number} rx The radius of the ellipse on the X axis.
 * @param {Number} ry The radius of the ellipse on the Y axis.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-ellipse)
 */
export declare const makeEllipse: ({ rx, ry }: MakeEllipseOptions) => ShapeInfo;
