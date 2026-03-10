"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.PathInternals = exports.warpPath = exports.translatePath = exports.serializeInstructions = exports.scalePath = exports.reversePath = exports.resetPath = exports.reduceInstructions = exports.parsePath = exports.normalizePath = exports.interpolatePath = exports.getTangentAtLength = exports.getSubpaths = exports.getPointAtLength = exports.getLength = exports.getInstructionIndexAtLength = exports.getBoundingBox = exports.extendViewBox = exports.evolvePath = exports.cutPath = void 0;
const cut_path_1 = require("./cut-path");
const debug_path_1 = require("./debug-path");
const get_bounding_box_1 = require("./get-bounding-box");
const cut_path_2 = require("./cut-path");
Object.defineProperty(exports, "cutPath", { enumerable: true, get: function () { return cut_path_2.cutPath; } });
const evolve_path_1 = require("./evolve-path");
Object.defineProperty(exports, "evolvePath", { enumerable: true, get: function () { return evolve_path_1.evolvePath; } });
const extend_viewbox_1 = require("./extend-viewbox");
Object.defineProperty(exports, "extendViewBox", { enumerable: true, get: function () { return extend_viewbox_1.extendViewBox; } });
const get_bounding_box_2 = require("./get-bounding-box");
Object.defineProperty(exports, "getBoundingBox", { enumerable: true, get: function () { return get_bounding_box_2.getBoundingBox; } });
const get_instruction_index_at_length_1 = require("./get-instruction-index-at-length");
Object.defineProperty(exports, "getInstructionIndexAtLength", { enumerable: true, get: function () { return get_instruction_index_at_length_1.getInstructionIndexAtLength; } });
const get_length_1 = require("./get-length");
Object.defineProperty(exports, "getLength", { enumerable: true, get: function () { return get_length_1.getLength; } });
const get_point_at_length_1 = require("./get-point-at-length");
Object.defineProperty(exports, "getPointAtLength", { enumerable: true, get: function () { return get_point_at_length_1.getPointAtLength; } });
const get_subpaths_1 = require("./get-subpaths");
Object.defineProperty(exports, "getSubpaths", { enumerable: true, get: function () { return get_subpaths_1.getSubpaths; } });
const get_tangent_at_length_1 = require("./get-tangent-at-length");
Object.defineProperty(exports, "getTangentAtLength", { enumerable: true, get: function () { return get_tangent_at_length_1.getTangentAtLength; } });
const interpolate_path_1 = require("./interpolate-path/interpolate-path");
Object.defineProperty(exports, "interpolatePath", { enumerable: true, get: function () { return interpolate_path_1.interpolatePath; } });
const normalize_path_1 = require("./normalize-path");
Object.defineProperty(exports, "normalizePath", { enumerable: true, get: function () { return normalize_path_1.normalizePath; } });
const parse_path_1 = require("./parse-path");
Object.defineProperty(exports, "parsePath", { enumerable: true, get: function () { return parse_path_1.parsePath; } });
const reduce_instructions_1 = require("./reduce-instructions");
Object.defineProperty(exports, "reduceInstructions", { enumerable: true, get: function () { return reduce_instructions_1.reduceInstructions; } });
const reset_path_1 = require("./reset-path");
Object.defineProperty(exports, "resetPath", { enumerable: true, get: function () { return reset_path_1.resetPath; } });
const reverse_path_1 = require("./reverse-path");
Object.defineProperty(exports, "reversePath", { enumerable: true, get: function () { return reverse_path_1.reversePath; } });
const scale_path_1 = require("./scale-path");
Object.defineProperty(exports, "scalePath", { enumerable: true, get: function () { return scale_path_1.scalePath; } });
const serialize_instructions_1 = require("./serialize-instructions");
Object.defineProperty(exports, "serializeInstructions", { enumerable: true, get: function () { return serialize_instructions_1.serializeInstructions; } });
const translate_path_1 = require("./translate-path");
Object.defineProperty(exports, "translatePath", { enumerable: true, get: function () { return translate_path_1.translatePath; } });
const warp_path_1 = require("./warp-path");
Object.defineProperty(exports, "warpPath", { enumerable: true, get: function () { return warp_path_1.warpPath; } });
exports.PathInternals = {
    getBoundingBoxFromInstructions: get_bounding_box_1.getBoundingBoxFromInstructions,
    debugPath: debug_path_1.debugPath,
    cutPath: cut_path_1.cutPath,
};
