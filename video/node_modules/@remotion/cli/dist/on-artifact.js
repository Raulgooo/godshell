"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.handleOnArtifact = void 0;
const fs_1 = require("fs");
const path_1 = __importDefault(require("path"));
const handleOnArtifact = ({ artifactState, onProgress, compositionId, }) => {
    const initialProgress = { ...artifactState };
    const onArtifact = (artifact) => {
        // It would be nice in the future to customize the artifact output destination
        const relativeOutputDestination = path_1.default.join('out', compositionId, artifact.filename.replace('/', path_1.default.sep));
        const defaultOutName = path_1.default.join(process.cwd(), relativeOutputDestination);
        if (!(0, fs_1.existsSync)(path_1.default.dirname(defaultOutName))) {
            (0, fs_1.mkdirSync)(path_1.default.dirname(defaultOutName), {
                recursive: true,
            });
        }
        const alreadyExisted = (0, fs_1.existsSync)(defaultOutName);
        (0, fs_1.writeFileSync)(defaultOutName, artifact.content);
        initialProgress.received.push({
            absoluteOutputDestination: defaultOutName,
            filename: artifact.filename,
            sizeInBytes: artifact.content.length,
            alreadyExisted,
            relativeOutputDestination,
        });
    };
    onProgress(initialProgress);
    return {
        onArtifact,
    };
};
exports.handleOnArtifact = handleOnArtifact;
