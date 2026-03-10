/**
 * Runs de Casteljau's algorithm enough times to produce the desired number of segments.
 *
 * @param {Number[][]} points Array of [x,y] points for de Casteljau (the initial segment to split)
 * @param {Number} segmentCount Number of segments to split the original into
 * @return {Number[][][]} Array of segments
 */
export declare function splitCurveAsPoints(points: number[][], segmentCount?: number): number[][][];
