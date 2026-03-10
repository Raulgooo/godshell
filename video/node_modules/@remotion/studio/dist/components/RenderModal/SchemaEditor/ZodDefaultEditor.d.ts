import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodDefaultEditor: React.FC<{
    readonly showSaveButton: boolean;
    readonly jsonPath: JSONPath;
    readonly value: unknown;
    readonly defaultValue: unknown;
    readonly schema: z.ZodTypeAny;
    readonly setValue: UpdaterFunction<unknown>;
    readonly onSave: UpdaterFunction<unknown>;
    readonly onRemove: null | (() => void);
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
