"use strict";
/**
 * https://github.com/Pomax/svg-path-reverse
 *
 * This code is in the public domain, except in jurisdictions that do
 * not recognise the public domain, where this code is MIT licensed.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.reversePath = void 0;
const construct_1 = require("./helpers/construct");
const normalize_path_1 = require("./normalize-path");
const parse_path_1 = require("./parse-path");
const reduce_instructions_1 = require("./reduce-instructions");
const serialize_instructions_1 = require("./serialize-instructions");
function reverseNormalizedPath(instructions) {
    const reversed = [];
    let nextX = 0;
    let nextY = 0;
    for (const term of instructions) {
        if (term.type === 'A') {
            reversed.unshift({
                type: 'A',
                largeArcFlag: term.largeArcFlag,
                rx: term.rx,
                ry: term.ry,
                xAxisRotation: term.xAxisRotation,
                sweepFlag: !term.sweepFlag,
                x: nextX,
                y: nextY,
            });
        }
        else if (term.type === 'C') {
            reversed.unshift({
                type: 'C',
                cp1x: term.cp2x,
                cp1y: term.cp2y,
                cp2x: term.cp1x,
                cp2y: term.cp1y,
                x: nextX,
                y: nextY,
            });
        }
        else if (term.type === 'Q') {
            reversed.unshift({
                type: 'Q',
                cpx: term.cpx,
                cpy: term.cpy,
                x: nextX,
                y: nextY,
            });
        }
        else if (term.type === 'L') {
            reversed.unshift({
                type: 'L',
                x: nextX,
                y: nextY,
            });
            // Do nothing
        }
        else if (term.type === 'M') {
            // Do nothing
        }
        else if (term.type === 'Z') {
            // Do nothing
        }
        else {
            throw new Error('unnormalized instruction ' + term.type);
        }
        if (term.type !== 'Z') {
            nextX = term.x;
            nextY = term.y;
        }
    }
    reversed.unshift({
        type: 'M',
        x: nextX,
        y: nextY,
    });
    let revstring = (0, serialize_instructions_1.serializeInstructions)(reversed);
    if (instructions[instructions.length - 1].type === 'Z')
        revstring += 'Z';
    revstring = revstring.replace(/M M/g, 'Z M');
    return revstring;
}
/*
 * @description Reverses a path so the end and start are switched.
 * @see [Documentation](https://www.remotion.dev/docs/paths/reverse-path)
 */
const reversePath = (path) => {
    const parsed = (0, parse_path_1.parsePath)(path);
    const normalized = (0, normalize_path_1.normalizeInstructions)(parsed);
    const reduced = (0, reduce_instructions_1.reduceInstructions)(normalized);
    const { segments } = (0, construct_1.constructFromInstructions)(reduced);
    return segments
        .map((spath) => {
        return reverseNormalizedPath(spath);
    })
        .join(' ')
        .replace(/ +/g, ' ')
        .trim();
};
exports.reversePath = reversePath;
