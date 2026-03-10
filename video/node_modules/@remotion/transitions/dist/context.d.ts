import React from 'react';
type EnteringState = {
    enteringProgress: number;
};
type ExitingState = {
    exitingProgress: number;
};
export declare const EnteringContext: React.Context<EnteringState | null>;
export declare const ExitingContext: React.Context<ExitingState | null>;
export declare const WrapInEnteringProgressContext: React.FC<{
    readonly presentationProgress: number;
    readonly children: React.ReactNode;
}>;
export declare const WrapInExitingProgressContext: React.FC<{
    readonly presentationProgress: number;
    readonly children: React.ReactNode;
}>;
export {};
