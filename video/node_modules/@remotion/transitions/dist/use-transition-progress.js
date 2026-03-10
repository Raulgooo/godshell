"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.useTransitionProgress = void 0;
const react_1 = __importDefault(require("react"));
const context_1 = require("./context");
/*
 * @description A hook that can be used inside a child of a <TransitionSeries.Sequence> to get the progress of the transition to directly manipulate the objects inside the scene.
 * @see [Documentation](https://www.remotion.dev/docs/transitions/use-transition-progress)
 */
const useTransitionProgress = () => {
    var _a, _b;
    const entering = react_1.default.useContext(context_1.EnteringContext);
    const exiting = react_1.default.useContext(context_1.ExitingContext);
    if (!entering && !exiting) {
        return {
            isInTransitionSeries: false,
            entering: 1,
            exiting: 0,
        };
    }
    return {
        isInTransitionSeries: true,
        entering: (_a = entering === null || entering === void 0 ? void 0 : entering.enteringProgress) !== null && _a !== void 0 ? _a : 1,
        exiting: (_b = exiting === null || exiting === void 0 ? void 0 : exiting.exitingProgress) !== null && _b !== void 0 ? _b : 0,
    };
};
exports.useTransitionProgress = useTransitionProgress;
