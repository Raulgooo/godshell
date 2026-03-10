export type TNonceContext = {
    getNonce: () => number;
};
type NonceHistoryItem = [number, number];
export type NonceHistory = NonceHistoryItem[];
export type NonceHistoryContext = {
    get: () => NonceHistory;
};
export declare const NonceContext: import("react").Context<TNonceContext>;
export declare const useNonce: () => NonceHistoryContext;
export {};
