import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodBooleanEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: boolean;
    readonly setValue: UpdaterFunction<boolean>;
    readonly defaultValue: boolean;
    readonly onSave: UpdaterFunction<boolean>;
    readonly onRemove: null | (() => void);
    readonly showSaveButton: boolean;
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
