import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodDateEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: Date;
    readonly defaultValue: Date;
    readonly setValue: UpdaterFunction<Date>;
    readonly onSave: UpdaterFunction<Date>;
    readonly onRemove: null | (() => void);
    readonly showSaveButton: boolean;
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
