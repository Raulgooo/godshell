"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.RenderSvg = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const react_1 = __importStar(require("react"));
const react_dom_1 = require("react-dom");
const does_react_support_canary_1 = require("../utils/does-react-support-canary");
const RenderSvg = ({ width, height, path, style, pathStyle, transformOrigin, debug, instructions, ...props }) => {
    const actualStyle = (0, react_1.useMemo)(() => {
        return {
            overflow: 'visible',
            ...(style !== null && style !== void 0 ? style : {}),
        };
    }, [style]);
    const actualPathStyle = (0, react_1.useMemo)(() => {
        return {
            transformBox: 'fill-box',
            ...(pathStyle !== null && pathStyle !== void 0 ? pathStyle : {}),
        };
    }, [pathStyle]);
    const reactSupportsTransformOrigin = (0, does_react_support_canary_1.doesReactSupportTransformOriginProperty)(react_dom_1.version);
    return (jsx_runtime_1.jsxs("svg", { width: width, height: height, viewBox: `0 0 ${width} ${height}`, xmlns: "http://www.w3.org/2000/svg", style: actualStyle, children: [
            jsx_runtime_1.jsx("path", { ...(reactSupportsTransformOrigin
                    ? {
                        transformOrigin,
                    }
                    : {
                        'transform-origin': transformOrigin,
                    }), d: path, style: actualPathStyle, ...props }), debug
                ? instructions.map((i, index) => {
                    if (i.type === 'C') {
                        const prevInstruction = index === 0
                            ? instructions[instructions.length - 1]
                            : instructions[index - 1];
                        if (prevInstruction.type === 'V' ||
                            prevInstruction.type === 'H' ||
                            prevInstruction.type === 'a' ||
                            prevInstruction.type === 'Z' ||
                            prevInstruction.type === 't' ||
                            prevInstruction.type === 'q' ||
                            prevInstruction.type === 'l' ||
                            prevInstruction.type === 'c' ||
                            prevInstruction.type === 'm' ||
                            prevInstruction.type === 'h' ||
                            prevInstruction.type === 's' ||
                            prevInstruction.type === 'v') {
                            return null;
                        }
                        const prevX = prevInstruction.x;
                        const prevY = prevInstruction.y;
                        return (
                        // eslint-disable-next-line react/no-array-index-key
                        jsx_runtime_1.jsxs(react_1.default.Fragment, { children: [
                                jsx_runtime_1.jsx("path", { d: `M ${prevX} ${prevY} ${i.cp1x} ${i.cp1y}`, strokeWidth: 2, stroke: "rgba(0, 0, 0, 0.4)" }), jsx_runtime_1.jsx("path", { d: `M ${i.x} ${i.y} ${i.cp2x} ${i.cp2y}`, strokeWidth: 2, stroke: "rgba(0, 0, 0, 0.4)" }), jsx_runtime_1.jsx("circle", { cx: i.cp1x, cy: i.cp1y, r: 3, fill: "white", strokeWidth: 2, stroke: "black" }), jsx_runtime_1.jsx("circle", { cx: i.cp2x, cy: i.cp2y, r: 3, strokeWidth: 2, fill: "white", stroke: "black" })
                            ] }, index));
                    }
                    return null;
                })
                : null] }));
};
exports.RenderSvg = RenderSvg;
