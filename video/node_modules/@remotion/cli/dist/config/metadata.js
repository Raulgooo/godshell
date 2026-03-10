"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getMetadata = exports.setMetadata = void 0;
let specifiedMetadata;
const setMetadata = (metadata) => {
    specifiedMetadata = metadata;
};
exports.setMetadata = setMetadata;
const getMetadata = () => {
    return specifiedMetadata;
};
exports.getMetadata = getMetadata;
