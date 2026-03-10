"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ready = ready;
function ready(context, callback) {
    if (context.state) {
        callback(context.stats);
        return;
    }
    context.callbacks.push(callback);
}
