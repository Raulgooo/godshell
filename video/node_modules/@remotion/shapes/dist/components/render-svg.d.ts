import type { Instruction } from '@remotion/paths';
import React from 'react';
export type AllShapesProps = Omit<React.SVGProps<SVGPathElement>, 'width' | 'height' | 'd'> & {
    readonly debug?: boolean;
    readonly pathStyle?: React.CSSProperties;
};
export declare const RenderSvg: ({ width, height, path, style, pathStyle, transformOrigin, debug, instructions, ...props }: {
    readonly width: number;
    readonly height: number;
    readonly path: string;
    readonly instructions: Instruction[];
    readonly transformOrigin: string;
} & Omit<React.SVGProps<SVGPathElement>, "d" | "height" | "width"> & {
    readonly debug?: boolean | undefined;
    readonly pathStyle?: React.CSSProperties | undefined;
}) => import("react/jsx-runtime").JSX.Element;
