import type { PackageManager } from '@remotion/studio-shared';
export declare const getInstallCommand: ({ manager, packages, version, additionalArgs, }: {
    manager: PackageManager;
    packages: string[];
    version: string;
    additionalArgs: string[];
}) => string[];
