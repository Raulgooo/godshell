"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.doesReactSupportTransformOriginProperty = void 0;
const doesReactSupportTransformOriginProperty = (version) => {
    if (version.includes('canary') || version.includes('experimental')) {
        const last8Chars = parseInt(version.slice(-8), 10);
        return last8Chars > 20230209;
    }
    const [major] = version.split('.').map(Number);
    return major > 18;
};
exports.doesReactSupportTransformOriginProperty = doesReactSupportTransformOriginProperty;
