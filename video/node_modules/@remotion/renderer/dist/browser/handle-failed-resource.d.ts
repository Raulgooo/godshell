import type { LogLevel } from '../log-level';
import type { HTTPRequest } from './HTTPRequest';
import type { LoadingFailedEvent, ResponseReceivedExtraInfoEvent } from './devtools-types';
export declare const handleFailedResource: ({ extraInfo, logLevel, indent, request, event, }: {
    extraInfo: ResponseReceivedExtraInfoEvent[];
    logLevel: LogLevel;
    indent: boolean;
    request: HTTPRequest;
    event: LoadingFailedEvent;
}) => void;
