import React from 'react';
import type { AnyZodObject, z } from 'zod';
export declare const SchemaEditor: React.FC<{
    readonly schema: AnyZodObject;
    readonly unsavedDefaultProps: Record<string, unknown>;
    readonly setValue: React.Dispatch<React.SetStateAction<Record<string, unknown>>>;
    readonly zodValidationResult: z.SafeParseReturnType<unknown, unknown>;
    readonly savedDefaultProps: Record<string, unknown>;
    readonly onSave: (updater: (oldState: Record<string, unknown>) => Record<string, unknown>) => void;
    readonly showSaveButton: boolean;
    readonly saving: boolean;
    readonly saveDisabledByParent: boolean;
}>;
