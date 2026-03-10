"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.setHeaderForResponse = setHeaderForResponse;
exports.send = send;
function setHeaderForResponse(res, name, value) {
    res.setHeader(name, typeof value === 'number' ? String(value) : value);
}
function send(req, res, bufferOtStream, byteLength) {
    if (typeof bufferOtStream === 'string' || Buffer.isBuffer(bufferOtStream)) {
        res.end(bufferOtStream);
        return;
    }
    setHeaderForResponse(res, 'Content-Length', byteLength);
    if (req.method === 'HEAD') {
        res.end();
        return;
    }
    bufferOtStream.pipe(res);
}
