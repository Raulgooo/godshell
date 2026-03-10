/*!
 * serve-static
 * Copyright(c) 2010 Sencha Inc.
 * Copyright(c) 2011 TJ Holowaychuk
 * Copyright(c) 2014-2016 Douglas Christopher Wilson
 * MIT Licensed
 */
import type { IncomingMessage, ServerResponse } from 'node:http';
export declare const serveStatic: ({ root, path, req, res, allowOutsidePublicFolder, }: {
    root: string;
    path: string;
    req: IncomingMessage;
    res: ServerResponse;
    allowOutsidePublicFolder: boolean;
}) => Promise<void>;
