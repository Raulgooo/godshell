"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.useSchema = void 0;
/* eslint-disable react-hooks/rules-of-hooks */
const react_1 = require("react");
const get_effective_visual_mode_value_js_1 = require("./get-effective-visual-mode-value.js");
const SequenceManager_js_1 = require("./SequenceManager.js");
const use_remotion_environment_js_1 = require("./use-remotion-environment.js");
const useSchema = (schema, currentValue) => {
    const env = (0, use_remotion_environment_js_1.useRemotionEnvironment)();
    const earlyReturn = (0, react_1.useMemo)(() => {
        if (!env.isStudio || env.isReadOnlyStudio) {
            return {
                controls: undefined,
                values: (currentValue !== null && currentValue !== void 0 ? currentValue : {}),
            };
        }
        return undefined;
    }, [env.isStudio, env.isReadOnlyStudio, currentValue]);
    if (earlyReturn) {
        return earlyReturn;
    }
    // Intentional conditional hook call, useRemotionEnvironment is stable.
    const [overrideId] = (0, react_1.useState)(() => String(Math.random()));
    const { visualModeEnabled, dragOverrides: overrides, codeValues, } = (0, react_1.useContext)(SequenceManager_js_1.VisualModeOverridesContext);
    const controls = (0, react_1.useMemo)(() => {
        if (!visualModeEnabled) {
            return undefined;
        }
        if (schema === null || currentValue === null) {
            return undefined;
        }
        return {
            schema,
            currentValue,
            overrideId,
        };
    }, [schema, currentValue, overrideId, visualModeEnabled]);
    return (0, react_1.useMemo)(() => {
        var _a, _b, _c;
        if (controls === undefined ||
            currentValue === null ||
            schema === null ||
            !visualModeEnabled) {
            return {
                controls: undefined,
                values: (currentValue !== null && currentValue !== void 0 ? currentValue : {}),
            };
        }
        const overrideValues = (_a = overrides[overrideId]) !== null && _a !== void 0 ? _a : {};
        const propStatus = codeValues[overrideId];
        const currentValueKeys = Object.keys(currentValue);
        const keysToUpdate = new Set(currentValueKeys).values();
        const merged = {};
        // Apply code values over runtime values, falling back to schema default
        for (const key of keysToUpdate) {
            const codeValueStatus = (_b = propStatus === null || propStatus === void 0 ? void 0 : propStatus[key]) !== null && _b !== void 0 ? _b : null;
            merged[key] = (0, get_effective_visual_mode_value_js_1.getEffectiveVisualModeValue)({
                codeValue: codeValueStatus,
                runtimeValue: currentValue[key],
                dragOverrideValue: overrideValues[key],
                defaultValue: (_c = schema[key]) === null || _c === void 0 ? void 0 : _c.default,
                shouldResortToDefaultValueIfUndefined: false,
            });
        }
        return {
            controls,
            values: merged,
        };
    }, [
        controls,
        currentValue,
        overrideId,
        overrides,
        codeValues,
        schema,
        visualModeEnabled,
    ]);
};
exports.useSchema = useSchema;
