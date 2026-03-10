import type { LogLevel } from '../log-level';
export declare const logLevelOption: {
    cliFlag: "log";
    name: string;
    ssrName: string;
    description: () => import("react/jsx-runtime").JSX.Element;
    docLink: string;
    getValue: ({ commandLine }: {
        commandLine: Record<string, unknown>;
    }) => {
        value: LogLevel;
        source: string;
    };
    setConfig: (newLogLevel: LogLevel) => void;
    type: LogLevel;
};
