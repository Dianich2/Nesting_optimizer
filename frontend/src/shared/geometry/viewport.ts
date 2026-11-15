export interface Viewport { x: number; y: number; width: number; height: number }

export function fitAspect(view: Viewport, aspect: number): Viewport {
    if (!Number.isFinite(aspect) || aspect <= 0) return view
    const width = Math.max(view.width, view.height * aspect)
    const height = width / aspect
    return { x: view.x + (view.width - width) / 2, y: view.y + (view.height - height) / 2, width, height }
}

export function zoomView(view: Viewport, factor: number, anchor = { x: view.x + view.width / 2, y: view.y + view.height / 2 }): Viewport {
    const width = view.width * factor
    if (!Number.isFinite(width) || width < 0.01 || width > 1e9) return view
    return { x: anchor.x + (view.x - anchor.x) * factor, y: anchor.y + (view.y - anchor.y) * factor, width, height: view.height * factor }
}
