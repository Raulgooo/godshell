"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.Polygon = void 0;
const jsx_runtime_1 = require("react/jsx-runtime");
const make_polygon_1 = require("../utils/make-polygon");
const render_svg_1 = require("./render-svg");
const Polygon = ({ points, radius, cornerRadius, edgeRoundness, ...props }) => {
    return (jsx_runtime_1.jsx(render_svg_1.RenderSvg, { ...(0, make_polygon_1.makePolygon)({
            points,
            cornerRadius,
            edgeRoundness,
            radius,
        }), ...props }));
};
exports.Polygon = Polygon;
