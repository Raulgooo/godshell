"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getInstructionIndexAtLength = exports.getInstructionIndexAtLengthFromConstructed = void 0;
const construct_1 = require("./helpers/construct");
const getInstructionIndexAtLengthFromConstructed = (constructed, fractionLength) => {
    if (fractionLength < 0) {
        throw new Error('Length less than 0 was passed');
    }
    if (fractionLength > constructed.totalLength) {
        fractionLength = constructed.totalLength;
    }
    let index = constructed.partialLengths.length - 1;
    while (constructed.partialLengths[index] >= fractionLength && index > 0) {
        index--;
    }
    return {
        lengthIntoInstruction: fractionLength - constructed.partialLengths[index],
        index,
    };
};
exports.getInstructionIndexAtLengthFromConstructed = getInstructionIndexAtLengthFromConstructed;
/*
 * @description Gets the index of the instruction and the part length into the instruction at a specified length along an SVG path.
 * @see [Documentation](https://www.remotion.dev/docs/paths/get-instruction-index-at-length)
 */
const getInstructionIndexAtLength = (path, length) => {
    const constructed = (0, construct_1.construct)(path);
    if (length > constructed.totalLength) {
        throw new Error(`A length of ${length} was passed to getInstructionIndexAtLength() but the total length of the path is only ${constructed.totalLength}`);
    }
    return (0, exports.getInstructionIndexAtLengthFromConstructed)(constructed, length);
};
exports.getInstructionIndexAtLength = getInstructionIndexAtLength;
