"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.overrideWebpackConfig = exports.getWebpackOverrideFn = exports.defaultOverrideFunction = void 0;
const defaultOverrideFunction = (config) => config;
exports.defaultOverrideFunction = defaultOverrideFunction;
let overrideFn = exports.defaultOverrideFunction;
const getWebpackOverrideFn = () => {
    return overrideFn;
};
exports.getWebpackOverrideFn = getWebpackOverrideFn;
const overrideWebpackConfig = (fn) => {
    const prevOverride = overrideFn;
    overrideFn = async (c) => fn(await prevOverride(c));
};
exports.overrideWebpackConfig = overrideWebpackConfig;
