import type { ShapeInfo } from './shape-info';
export type MakeHeartProps = {
    height: number;
    aspectRatio?: number;
    bottomRoundnessAdjustment?: number;
    depthAdjustment?: number;
};
/**
 * @description Generates a heart SVG path.
 * @param {Number} size The size of the heart.
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-heart)
 */
export declare const makeHeart: ({ height, aspectRatio, bottomRoundnessAdjustment, depthAdjustment, }: MakeHeartProps) => ShapeInfo;
