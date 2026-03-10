import type { Instruction } from './helpers/types';
export declare const translateSegments: (segments: Instruction[], x: number, y: number) => Instruction[];
export declare const translatePath: (path: string, x: number, y: number) => string;
