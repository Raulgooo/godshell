"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getEffectiveVisualModeValue = void 0;
const getEffectiveVisualModeValue = ({ codeValue, runtimeValue, dragOverrideValue, defaultValue, shouldResortToDefaultValueIfUndefined = false, }) => {
    if (dragOverrideValue !== undefined) {
        return dragOverrideValue;
    }
    if (!codeValue) {
        return runtimeValue;
    }
    if (!codeValue.canUpdate) {
        return runtimeValue;
    }
    if (codeValue.codeValue === undefined &&
        shouldResortToDefaultValueIfUndefined) {
        return defaultValue;
    }
    return codeValue.codeValue;
};
exports.getEffectiveVisualModeValue = getEffectiveVisualModeValue;
