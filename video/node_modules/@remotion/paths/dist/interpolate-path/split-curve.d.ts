import type { CInstruction, LInstruction, QInstruction, ReducedInstruction } from '../helpers/types';
/**
 * Convert command objects to arrays of points, run de Casteljau's algorithm on it
 * to split into to the desired number of segments.
 *
 * @param {Object} commandStart The start command object
 * @param {Object} instructionEnd The end command object
 * @param {Number} segmentCount The number of segments to create
 * @return {Object[]} An array of commands representing the segments in sequence
 */
export declare const splitCurveInstructions: (instructionStartX: number, instructionStartY: number, instructionEnd: CInstruction | LInstruction | QInstruction, segmentCount: number) => ReducedInstruction[];
