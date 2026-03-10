"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getPointAtLength = void 0;
const get_instruction_index_at_length_1 = require("./get-instruction-index-at-length");
const construct_1 = require("./helpers/construct");
/*
 * @description Gets the coordinates of a point which is on an SVG path.
 * @see [Documentation](https://www.remotion.dev/docs/paths/get-point-at-length)
 */
const getPointAtLength = (path, length) => {
    const constructed = (0, construct_1.construct)(path);
    const fractionPart = (0, get_instruction_index_at_length_1.getInstructionIndexAtLengthFromConstructed)(constructed, length);
    const functionAtPart = constructed.functions[fractionPart.index + 1];
    if (functionAtPart) {
        return functionAtPart.getPointAtLength(fractionPart.lengthIntoInstruction);
    }
    if (constructed.initialPoint) {
        return constructed.initialPoint;
    }
    throw new Error('Wrong function at this part.');
};
exports.getPointAtLength = getPointAtLength;
