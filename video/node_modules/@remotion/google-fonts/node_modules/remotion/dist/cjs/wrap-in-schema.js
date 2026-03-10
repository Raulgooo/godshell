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
exports.wrapInSchema = void 0;
const react_1 = __importStar(require("react"));
const use_remotion_environment_js_1 = require("./use-remotion-environment.js");
const use_schema_js_1 = require("./use-schema.js");
const getNestedValue = (obj, key) => {
    const parts = key.split('.');
    let current = obj;
    for (const part of parts) {
        if (current === null ||
            current === undefined ||
            typeof current !== 'object')
            return undefined;
        current = current[part];
    }
    return current;
};
const mergeValues = (props, values, schemaKeys) => {
    const merged = { ...props };
    for (const key of schemaKeys) {
        const value = values[key];
        const parts = key.split('.');
        if (parts.length === 1) {
            merged[key] = value;
            continue;
        }
        // For dot-notation keys like 'style.opacity',
        // clone and set the nested path
        let current = merged;
        for (let i = 0; i < parts.length - 1; i++) {
            const part = parts[i];
            if (typeof current[part] === 'object' && current[part] !== null) {
                current[part] = { ...current[part] };
            }
            else {
                current[part] = {};
            }
            current = current[part];
        }
        current[parts[parts.length - 1]] = value;
    }
    return merged;
};
const wrapInSchema = (Component, schema) => {
    const schemaKeys = Object.keys(schema);
    const Wrapped = (0, react_1.forwardRef)((props, ref) => {
        const env = (0, use_remotion_environment_js_1.useRemotionEnvironment)();
        if (!env.isStudio ||
            env.isReadOnlyStudio ||
            env.isRendering ||
            !process.env.EXPERIMENTAL_VISUAL_MODE_ENABLED) {
            return react_1.default.createElement(Component, {
                ...props,
                controls: null,
                ref,
            });
        }
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const schemaInput = (0, react_1.useMemo)(() => {
            const input = {};
            for (const key of schemaKeys) {
                input[key] = getNestedValue(props, key);
            }
            return input;
        }, 
        // eslint-disable-next-line react-hooks/exhaustive-deps
        schemaKeys.map((key) => getNestedValue(props, key)));
        // eslint-disable-next-line react-hooks/rules-of-hooks
        const { controls, values } = (0, use_schema_js_1.useSchema)(schema, schemaInput);
        const mergedProps = mergeValues(props, values, schemaKeys);
        return react_1.default.createElement(Component, {
            ...mergedProps,
            controls,
            ref,
        });
    });
    Wrapped.displayName = `wrapInSchema(${Component.displayName || Component.name || 'Component'})`;
    return Wrapped;
};
exports.wrapInSchema = wrapInSchema;
