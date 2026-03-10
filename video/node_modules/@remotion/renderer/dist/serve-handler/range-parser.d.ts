/*!
 * range-parser
 * Copyright(c) 2012-2014 TJ Holowaychuk
 * Copyright(c) 2015-2016 Douglas Christopher Wilson
 * MIT Licensed
 */
export declare const rangeParser: (size: number, str: string) => {
    type: string;
    ranges: {
        start: number;
        end: number;
    }[];
} | -1 | -2;
