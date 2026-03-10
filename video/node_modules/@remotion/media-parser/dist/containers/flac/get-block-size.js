"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getBlockSize = void 0;
const getBlockSize = (iterator) => {
    const bits = iterator.getBits(4);
    if (bits === 0b0000) {
        throw new Error('Reserved block size');
    }
    if (bits === 0b0001) {
        return 192;
    }
    if (bits >= 0b0010 && bits <= 0b0101) {
        return 144 * 2 ** bits;
    }
    if (bits === 0b0110) {
        return 'uncommon-u8';
    }
    if (bits === 0b0111) {
        return 'uncommon-u16';
    }
    if (bits >= 0b1000 && bits <= 0b1111) {
        return 2 ** bits;
    }
    throw new Error('Invalid block size');
};
exports.getBlockSize = getBlockSize;
