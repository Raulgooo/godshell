import React from 'react';
import type { MakePolygonProps } from '../utils/make-polygon';
import type { AllShapesProps } from './render-svg';
export type PolygonProps = MakePolygonProps & Omit<AllShapesProps, 'points'>;
export declare const Polygon: React.FC<PolygonProps>;
