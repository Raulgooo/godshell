"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.debugPath = void 0;
const normalize_path_1 = require("./normalize-path");
const parse_path_1 = require("./parse-path");
const serialize_instructions_1 = require("./serialize-instructions");
const debugPath = (d) => {
    const instructions = (0, normalize_path_1.normalizeInstructions)((0, parse_path_1.parsePath)(d));
    return instructions
        .map((inst, i) => {
        if (inst.type === 'Z') {
            return null;
        }
        if (inst.type === 'H' || inst.type === 'V') {
            return null;
        }
        const topLeft = [inst.x - 5, inst.y - 5];
        const topRight = [inst.x + 5, inst.y - 5];
        const bottomLeft = [inst.x - 5, inst.y + 5];
        const bottomRight = [inst.x + 5, inst.y + 5];
        const triangle = [
            {
                type: 'M',
                x: topLeft[0],
                y: topLeft[1],
            },
            {
                type: 'L',
                x: topRight[0],
                y: topRight[1],
            },
            {
                type: 'L',
                x: bottomRight[0],
                y: bottomRight[1],
            },
            {
                type: 'L',
                x: bottomLeft[0],
                y: bottomLeft[1],
            },
            {
                type: 'Z',
            },
        ];
        return {
            d: (0, serialize_instructions_1.serializeInstructions)(triangle),
            color: i === instructions.length - 1
                ? 'red'
                : inst.type === 'M'
                    ? 'blue'
                    : 'green',
        };
    })
        .filter(Boolean);
};
exports.debugPath = debugPath;
