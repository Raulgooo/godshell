import type { VideoMetadata } from './compositor/payloads';
import type { LogLevel } from './log-level';
export { VideoMetadata } from './compositor/payloads';
/**
 * @deprecated Use `parseMedia()` instead: https://www.remotion.dev/docs/media-parser/parse-media
 */
export declare const getVideoMetadata: (videoSource: string, options?: {
    logLevel?: LogLevel;
    binariesDirectory?: string | null;
}) => Promise<VideoMetadata>;
