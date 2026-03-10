import type { ReducedInstruction } from '../helpers/types';
/**
 * Interpolate from A to B by extending A and B during interpolation to have
 * the same number of points. This allows for a smooth transition when they
 * have a different number of points.
 *
 * Ignores the `Z` command in paths unless both A and B end with it.
 *
 * This function works directly with arrays of command objects instead of with
 * path `d` strings (see interpolatePath for working with `d` strings).
 *
 * @param {Object[]} aCommandsInput Array of path commands
 * @param {Object[]} bCommandsInput Array of path commands
 * @returns {Function} Interpolation function that maps t ([0, 1]) to an array of path commands.
 */
export declare function interpolateInstructions(aCommandsInput: ReducedInstruction[], bCommandsInput: ReducedInstruction[]): (t: number) => ReducedInstruction[];
