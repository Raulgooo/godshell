"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.cutPath = void 0;
const cut_instruction_1 = require("./cut-instruction");
const reduced_analysis_1 = require("./helpers/reduced-analysis");
const parse_path_1 = require("./parse-path");
const reduce_instructions_1 = require("./reduce-instructions");
const serialize_instructions_1 = require("./serialize-instructions");
const cutPath = (d, length) => {
    const parsed = (0, parse_path_1.parsePath)(d);
    const reduced = (0, reduce_instructions_1.reduceInstructions)(parsed);
    const constructed = (0, reduced_analysis_1.conductAnalysis)(reduced);
    const newInstructions = [];
    let summedUpLength = 0;
    for (const segment of constructed) {
        for (const instructionAndInfo of segment.instructionsAndInfo) {
            if (summedUpLength + instructionAndInfo.length > length) {
                const remainingLength = length - summedUpLength;
                const progress = remainingLength / instructionAndInfo.length;
                // cut
                const cut = (0, cut_instruction_1.cutInstruction)({
                    instruction: instructionAndInfo.instruction,
                    lastPoint: instructionAndInfo.startPoint,
                    progress,
                });
                newInstructions.push(cut);
                return (0, serialize_instructions_1.serializeInstructions)(newInstructions);
            }
            summedUpLength += instructionAndInfo.length;
            newInstructions.push(instructionAndInfo.instruction);
            if (summedUpLength === length) {
                return (0, serialize_instructions_1.serializeInstructions)(newInstructions);
            }
        }
    }
    return (0, serialize_instructions_1.serializeInstructions)(newInstructions);
};
exports.cutPath = cutPath;
