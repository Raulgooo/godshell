import React from 'react';
import type { z } from 'zod';
export declare const ZodErrorMessages: React.FC<{
    readonly zodValidationResult: z.SafeParseReturnType<unknown, unknown>;
    readonly viewTab: 'schema' | 'json';
}>;
