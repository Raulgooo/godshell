import type { ReducedInstruction } from '../helpers/types';
/**
 * Convert segments represented as points back into a command object
 *
 * @param {Number[][]} points Array of [x,y] points: [start, control1, control2, ..., end]
 *   Represents a segment
 * @return {Object} A command object representing the segment.
 */
export declare function pointsToInstruction(points: number[][]): ReducedInstruction;
