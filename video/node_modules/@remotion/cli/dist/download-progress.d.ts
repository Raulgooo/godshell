import type { DownloadProgress } from '@remotion/studio-server';
export declare const getFileSizeDownloadBar: (downloaded: number) => string;
export declare const makeMultiDownloadProgress: (progresses: DownloadProgress[], totalFrames: number) => string | null;
