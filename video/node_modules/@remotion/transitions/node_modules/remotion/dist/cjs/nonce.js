"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.useNonce = exports.NonceContext = void 0;
const react_1 = require("react");
exports.NonceContext = (0, react_1.createContext)({
    getNonce: () => 0,
});
let fastRefreshNonce = 0;
try {
    if (typeof __webpack_module__ !== 'undefined') {
        if (__webpack_module__.hot) {
            __webpack_module__.hot.addStatusHandler((status) => {
                if (status === 'idle') {
                    fastRefreshNonce++;
                }
            });
        }
    }
}
catch (_a) {
    // __webpack_module__ may throw ReferenceError in some environments
}
const useNonce = () => {
    const context = (0, react_1.useContext)(exports.NonceContext);
    const nonce = context.getNonce();
    const nonceRef = (0, react_1.useRef)(nonce);
    nonceRef.current = nonce;
    const history = (0, react_1.useRef)([[fastRefreshNonce, nonce]]);
    const get = (0, react_1.useCallback)(() => {
        if (fastRefreshNonce !== history.current[history.current.length - 1][0]) {
            history.current = [
                ...history.current,
                [fastRefreshNonce, nonceRef.current],
            ];
        }
        return history.current;
    }, [history]);
    return (0, react_1.useMemo)(() => {
        return { get };
    }, [get]);
};
exports.useNonce = useNonce;
