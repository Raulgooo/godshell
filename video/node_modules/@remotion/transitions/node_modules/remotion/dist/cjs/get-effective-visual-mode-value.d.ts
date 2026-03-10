import type { CanUpdateSequencePropStatus } from './use-schema';
export declare const getEffectiveVisualModeValue: ({ codeValue, runtimeValue, dragOverrideValue, defaultValue, shouldResortToDefaultValueIfUndefined, }: {
    codeValue: CanUpdateSequencePropStatus | null;
    runtimeValue: unknown;
    dragOverrideValue: unknown;
    defaultValue: unknown;
    shouldResortToDefaultValueIfUndefined: boolean;
}) => unknown;
