export interface Viewport {
    x: number
    y: number
    width: number
    height: number
}

export function fitAspect(view: Viewport, aspect: number): Viewport {
    if (!Number.isFinite(aspect) || aspect <= 0) return view
    const width = Math.max(view.width, view.height * aspect)
    const height = width / aspect
    return {
        x: view.x + (view.width - width) / 2,
        y: view.y + (view.height - height) / 2,
        width,
        height,
    }
}

export function zoomView(
    view: Viewport,
    factor: number,
    anchor = { x: view.x + view.width / 2, y: view.y + view.height / 2 },
): Viewport {
    const width = view.width * factor
    if (!Number.isFinite(width) || width < 0.01 || width > 1e9) return view
    return {
        x: anchor.x + (view.x - anchor.x) * factor,
        y: anchor.y + (view.y - anchor.y) * factor,
        width,
        height: view.height * factor,
    }
}

export function screenToWorld(svg: SVGSVGElement, x: number, y: number) {
    const matrix = svg.getScreenCTM()
    if (!matrix) return null
    const point = new DOMPoint(x, y).matrixTransform(matrix.inverse())
    return Number.isFinite(point.x) && Number.isFinite(point.y) ? point : null
}

export function snapCoordinate(value: number, step: number, enabled: boolean): number {
    return enabled ? Math.round(value / step) * step : value
}

export function angleAround(
    center: { x: number; y: number },
    point: { x: number; y: number },
): number {
    return (Math.atan2(point.y - center.y, point.x - center.x) * 180) / Math.PI
}

export function normalizeAngle(degrees: number): number {
    return ((degrees % 360) + 360) % 360
}
