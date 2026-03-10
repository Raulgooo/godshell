export declare const makeHyperlink: ({ text, url, fallback, }: {
    text: string | ((clickInstruction: string) => string);
    url: string;
    fallback: string;
}) => string;
