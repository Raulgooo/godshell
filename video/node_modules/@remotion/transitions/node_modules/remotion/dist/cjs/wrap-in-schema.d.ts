import React from 'react';
import type { SequenceControls } from './CompositionManager.js';
import type { SequenceSchema } from './sequence-field-schema.js';
export declare const wrapInSchema: <S extends SequenceSchema, Props extends object>(Component: React.ComponentType<Props & {
    readonly controls: SequenceControls | undefined;
}>, schema: S) => React.ComponentType<Props>;
