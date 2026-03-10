"use strict";
// Copied from: https://github.com/rveciana/svg-path-properties
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeCubic = exports.makeQuadratic = void 0;
const bezier_functions_1 = require("./bezier-functions");
const makeQuadratic = ({ startX, startY, cpx, cpy, x, y, }) => {
    const a = { x: startX, y: startY };
    const b = { x: cpx, y: cpy };
    const c = { x, y };
    const length = (0, bezier_functions_1.getQuadraticArcLength)([a.x, b.x, c.x, 0], [a.y, b.y, c.y, 0], 1);
    const getTotalLength = () => {
        return length;
    };
    const getPointAtLength = (len) => {
        const xs = [a.x, b.x, c.x, 0];
        const xy = [a.y, b.y, c.y, 0];
        const t = (0, bezier_functions_1.t2length)({
            length: len,
            totalLength: length,
            func: (i) => (0, bezier_functions_1.getQuadraticArcLength)(xs, xy, i),
        });
        return (0, bezier_functions_1.quadraticPoint)(xs, xy, t);
    };
    const getTangentAtLength = (len) => {
        const xs = [a.x, b.x, c.x, 0];
        const xy = [a.y, b.y, c.y, 0];
        const t = (0, bezier_functions_1.t2length)({
            length: len,
            totalLength: length,
            func: (i) => (0, bezier_functions_1.getQuadraticArcLength)(xs, xy, i),
        });
        const derivative = (0, bezier_functions_1.quadraticDerivative)(xs, xy, t);
        const mdl = Math.sqrt(derivative.x * derivative.x + derivative.y * derivative.y);
        let tangent;
        if (mdl > 0) {
            tangent = { x: derivative.x / mdl, y: derivative.y / mdl };
        }
        else {
            tangent = { x: 0, y: 0 };
        }
        return tangent;
    };
    const getC = () => {
        return c;
    };
    return {
        getPointAtLength,
        getTangentAtLength,
        getTotalLength,
        getC,
        type: 'quadratic-bezier',
        getD: () => ({ x: 0, y: 0 }),
    };
};
exports.makeQuadratic = makeQuadratic;
const makeCubic = ({ startX, startY, cp1x, cp1y, cp2x, cp2y, x, y, }) => {
    const a = { x: startX, y: startY };
    const b = { x: cp1x, y: cp1y };
    const c = { x: cp2x, y: cp2y };
    const d = { x, y };
    const length = (0, bezier_functions_1.getCubicArcLength)({
        sx: [a.x, b.x, c.x, d.x],
        sy: [a.y, b.y, c.y, d.y],
        t: 1,
    });
    const getTotalLength = () => {
        return length;
    };
    const getPointAtLength = (len) => {
        const sx = [a.x, b.x, c.x, d.x];
        const sy = [a.y, b.y, c.y, d.y];
        const t = (0, bezier_functions_1.t2length)({
            length: len,
            totalLength: length,
            func: (i) => {
                return (0, bezier_functions_1.getCubicArcLength)({ sx, sy, t: i });
            },
        });
        return (0, bezier_functions_1.cubicPoint)(sx, sy, t);
    };
    const getTangentAtLength = (len) => {
        const xs = [a.x, b.x, c.x, d.x];
        const xy = [a.y, b.y, c.y, d.y];
        const t = (0, bezier_functions_1.t2length)({
            length: len,
            totalLength: length,
            func: (i) => (0, bezier_functions_1.getCubicArcLength)({ sx: xs, sy: xy, t: i }),
        });
        const derivative = (0, bezier_functions_1.cubicDerivative)(xs, xy, t);
        const mdl = Math.sqrt(derivative.x * derivative.x + derivative.y * derivative.y);
        let tangent;
        if (mdl > 0) {
            tangent = { x: derivative.x / mdl, y: derivative.y / mdl };
        }
        else {
            tangent = { x: 0, y: 0 };
        }
        return tangent;
    };
    const getC = () => {
        return c;
    };
    const getD = () => {
        return d;
    };
    return {
        getPointAtLength,
        getTangentAtLength,
        getTotalLength,
        getC,
        getD,
        type: 'cubic-bezier',
    };
};
exports.makeCubic = makeCubic;
