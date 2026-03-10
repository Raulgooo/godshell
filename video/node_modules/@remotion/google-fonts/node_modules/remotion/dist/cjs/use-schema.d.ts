import type { SequenceControls } from './CompositionManager.js';
import type { SchemaKeysRecord, SequenceSchema } from './sequence-field-schema.js';
export type CanUpdateSequencePropStatus = {
    canUpdate: true;
    codeValue: unknown;
} | {
    canUpdate: false;
    reason: 'computed';
};
export declare const useSchema: <S extends SequenceSchema, T extends SchemaKeysRecord<S>>(schema: S | null, currentValue: (T & Record<Exclude<keyof T, keyof S>, never>) | null) => {
    controls: SequenceControls | undefined;
    values: T;
};
