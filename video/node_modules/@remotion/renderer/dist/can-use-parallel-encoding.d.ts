import type { Codec } from './codec';
export declare const canUseParallelEncoding: (codec: Codec) => boolean;
export declare const getShouldUsePartitionedRendering: () => boolean;
