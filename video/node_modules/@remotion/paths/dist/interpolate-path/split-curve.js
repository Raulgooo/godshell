"use strict";
/*

Copied and adapted from https://github.com/pbeshai/d3-interpolate-path:
Copyright 2016, Peter Beshai
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of the author nor the names of contributors may be used to
  endorse or promote products derived from this software without specific prior
  written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/
Object.defineProperty(exports, "__esModule", { value: true });
exports.splitCurveInstructions = void 0;
const points_to_command_1 = require("./points-to-command");
const split_curve_as_points_1 = require("./split-curve-as-points");
/**
 * Convert command objects to arrays of points, run de Casteljau's algorithm on it
 * to split into to the desired number of segments.
 *
 * @param {Object} commandStart The start command object
 * @param {Object} instructionEnd The end command object
 * @param {Number} segmentCount The number of segments to create
 * @return {Object[]} An array of commands representing the segments in sequence
 */
const splitCurveInstructions = (instructionStartX, instructionStartY, instructionEnd, segmentCount) => {
    const points = [[instructionStartX, instructionStartY]];
    if (instructionEnd.type === 'Q') {
        points.push([instructionEnd.cpx, instructionEnd.cpy]);
    }
    if (instructionEnd.type === 'C') {
        points.push([instructionEnd.cp1x, instructionEnd.cp1y]);
        points.push([instructionEnd.cp2x, instructionEnd.cp2y]);
    }
    points.push([instructionEnd.x, instructionEnd.y]);
    return (0, split_curve_as_points_1.splitCurveAsPoints)(points, segmentCount).map((p) => {
        return (0, points_to_command_1.pointsToInstruction)(p);
    });
};
exports.splitCurveInstructions = splitCurveInstructions;
