import React from 'react';
import type { z } from 'zod';
import type { UpdaterFunction } from './ZodSwitch';
import type { JSONPath } from './zod-types';
export declare const ZodStaticFileEditor: React.FC<{
    readonly schema: z.ZodTypeAny;
    readonly jsonPath: JSONPath;
    readonly value: string;
    readonly defaultValue: string;
    readonly setValue: UpdaterFunction<string>;
    readonly onSave: (updater: (oldState: string) => string) => void;
    readonly showSaveButton: boolean;
    readonly onRemove: null | (() => void);
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
    readonly mayPad: boolean;
}>;
