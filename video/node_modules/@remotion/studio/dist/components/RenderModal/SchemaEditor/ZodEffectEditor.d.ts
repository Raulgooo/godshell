import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodEffectEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: unknown;
    readonly setValue: UpdaterFunction<unknown>;
    readonly defaultValue: unknown;
    readonly onSave: UpdaterFunction<unknown>;
    readonly showSaveButton: boolean;
    readonly onRemove: null | (() => void);
    readonly saving: boolean;
    readonly mayPad: boolean;
}>;
