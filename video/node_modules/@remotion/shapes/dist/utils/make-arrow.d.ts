import type { ShapeInfo } from './shape-info';
type ArrowDirection = 'right' | 'left' | 'up' | 'down';
export type MakeArrowProps = {
    length?: number;
    headWidth?: number;
    headLength?: number;
    shaftWidth?: number;
    direction?: ArrowDirection;
    cornerRadius?: number;
};
/**
 * @description Generates an SVG path for an arrow shape.
 * @param {Number} length The total length of the arrow along its direction axis. Default 300.
 * @param {Number} headWidth The width of the arrowhead at its widest point. Default 185.
 * @param {Number} headLength The length of the arrowhead portion. Default 120.
 * @param {Number} shaftWidth The width of the arrow shaft. Default 80.
 * @param {string} direction The direction the arrow points. Default 'right'.
 * @param {Number} cornerRadius Rounds the corner using an arc. Similar to CSS's border-radius.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-arrow)
 */
export declare const makeArrow: ({ length, headWidth, headLength, shaftWidth, direction, cornerRadius, }: MakeArrowProps) => ShapeInfo;
export {};
