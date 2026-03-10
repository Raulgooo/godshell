import type { Point, Properties } from './types';
export declare const makeQuadratic: ({ startX, startY, cpx, cpy, x, y, }: {
    startX: number;
    startY: number;
    cpx: number;
    cpy: number;
    x: number;
    y: number;
}) => Properties & {
    getC: () => Point;
    getD: () => Point;
};
export declare const makeCubic: ({ startX, startY, cp1x, cp1y, cp2x, cp2y, x, y, }: {
    startX: number;
    startY: number;
    cp1x: number;
    cp1y: number;
    cp2x: number;
    cp2y: number;
    x: number;
    y: number;
}) => Properties & {
    getC: () => Point;
    getD: () => Point;
};
