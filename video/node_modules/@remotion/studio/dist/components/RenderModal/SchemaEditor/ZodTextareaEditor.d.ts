import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodTextareaEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: string;
    readonly defaultValue: string;
    readonly setValue: UpdaterFunction<string>;
    readonly onSave: UpdaterFunction<string>;
    readonly onRemove: null | (() => void);
    readonly showSaveButton: boolean;
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
