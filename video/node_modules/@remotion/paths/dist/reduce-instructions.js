"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.reduceInstructions = void 0;
const remove_a_s_t_curves_1 = require("./helpers/remove-a-s-t-curves");
const normalize_path_1 = require("./normalize-path");
/*
 * @description Takes an array of Instruction's and reduces the amount of instruction types so that the path only consists of M, L, C, and Z instructions.
 * @see [Documentation](https://www.remotion.dev/docs/paths/reduce-instructions)
 */
const reduceInstructions = (instruction) => {
    const simplified = (0, normalize_path_1.normalizeInstructions)(instruction);
    return (0, remove_a_s_t_curves_1.removeATSHVQInstructions)(simplified);
};
exports.reduceInstructions = reduceInstructions;
