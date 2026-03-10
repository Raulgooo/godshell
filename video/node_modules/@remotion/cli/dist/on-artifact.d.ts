import type { OnArtifact } from '@remotion/renderer';
import type { ArtifactProgress } from '@remotion/studio-shared';
export declare const handleOnArtifact: ({ artifactState, onProgress, compositionId, }: {
    artifactState: ArtifactProgress;
    onProgress: (artifact: ArtifactProgress) => void;
    compositionId: string;
}) => {
    onArtifact: OnArtifact;
};
