import type { ReducedInstruction } from '../helpers/types';
/**
 * Extends an array of commandsToExtend to the length of the referenceCommands by
 * splitting segments until the number of commands match. Ensures all the actual
 * points of commandsToExtend are in the extended array.
 *
 * @param {Object[]} commandsToExtend The command object array to extend
 * @param {Object[]} referenceCommands The command object array to match in length
 * @return {Object[]} The extended commandsToExtend array
 */
export declare function extendInstruction(commandsToExtend: ReducedInstruction[], referenceCommands: ReducedInstruction[]): ReducedInstruction[];
