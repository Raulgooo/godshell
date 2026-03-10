import type { Instruction } from '@remotion/paths';
export type MakePieProps = {
    radius: number;
    progress: number;
    closePath?: boolean;
    counterClockwise?: boolean;
    rotation?: number;
};
/**
 * @description Generates a piece of pie SVG path.
 * @param {Number} radius The radius of the circle..
 * @param {Number} progress The percentage of the circle that is filled. 0 means fully empty, 1 means fully filled.
 * @param {Boolean} closePath If set to false, no path to the middle of the circle will be drawn, leading to an open arc. Default true.
 * @param {Boolean} counterClockwise If set, the circle gets filled counterclockwise instead of clockwise. Default false.
 * @param {Number} rotation Add rotation to the path. 0 means no rotation, Math.PI * 2 means 1 full clockwise rotation
 * @see [Documentation](https://www.remotion.dev/docs/shapes/make-pie)
 */
export declare const makePie: ({ progress, radius, closePath, counterClockwise, rotation, }: MakePieProps) => {
    height: number;
    width: number;
    path: string;
    instructions: Instruction[];
    transformOrigin: string;
};
