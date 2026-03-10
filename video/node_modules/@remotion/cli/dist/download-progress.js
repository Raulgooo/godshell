"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeMultiDownloadProgress = exports.getFileSizeDownloadBar = void 0;
const studio_server_1 = require("@remotion/studio-server");
const studio_shared_1 = require("@remotion/studio-shared");
const chalk_1 = require("./chalk");
const make_link_1 = require("./hyperlinks/make-link");
const make_progress_bar_1 = require("./make-progress-bar");
const progress_bar_1 = require("./progress-bar");
const truthy_1 = require("./truthy");
const getFileSizeDownloadBar = (downloaded) => {
    const desiredLength = (0, make_progress_bar_1.makeProgressBar)(0, true).length;
    return `${studio_server_1.StudioServerInternals.formatBytes(downloaded).padEnd(desiredLength, ' ')}`;
};
exports.getFileSizeDownloadBar = getFileSizeDownloadBar;
const makeMultiDownloadProgress = (progresses, totalFrames) => {
    if (progresses.length === 0) {
        return null;
    }
    const everyFileHasContentLength = progresses.every((p) => p.totalBytes !== null);
    const isDone = progresses.every((p) => p.progress === 1);
    const topRow = [
        (isDone ? `Downloaded assets` : 'Downloading assets').padEnd(progress_bar_1.LABEL_WIDTH, ' '),
        everyFileHasContentLength
            ? (0, make_progress_bar_1.makeProgressBar)(progresses.reduce((a, b) => a + b.progress, 0) /
                progresses.length, false)
            : (0, exports.getFileSizeDownloadBar)(progresses.reduce((a, b) => a + b.downloaded, 0)),
        `${progresses.length} file${progresses.length === 1 ? '' : 's'}`.padStart((0, progress_bar_1.getRightLabelWidth)(totalFrames), ' '),
    ]
        .filter(truthy_1.truthy)
        .join(' ');
    const downloadsToShow = progresses
        .filter((p) => p.progress !== 1)
        .slice(0, 2);
    return [
        topRow,
        ...downloadsToShow.map((toShow) => {
            const truncatedFileName = toShow.name.length >= 60
                ? toShow.name.substring(0, 57) + '...'
                : toShow.name;
            return chalk_1.chalk.gray(`↓ ${(0, studio_shared_1.formatBytes)(toShow.downloaded).padEnd(8, ' ')} ${(0, make_link_1.makeHyperlink)({ url: toShow.name, fallback: truncatedFileName, text: truncatedFileName })}`);
        }),
    ].join('\n');
};
exports.makeMultiDownloadProgress = makeMultiDownloadProgress;
