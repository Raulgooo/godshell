import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodNumberEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: number;
    readonly setValue: UpdaterFunction<number>;
    readonly defaultValue: number;
    readonly onSave: UpdaterFunction<number>;
    readonly onRemove: null | (() => void);
    readonly showSaveButton: boolean;
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
