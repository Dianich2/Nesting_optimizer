import { useEffect, useEffectEvent } from 'react'
import type { RefObject } from 'react'
import { screenToWorld, zoomView } from './viewport'
import type { Viewport } from './viewport'

export function useWheelZoom(
    ref: RefObject<SVGSVGElement | null>,
    view: Viewport,
    setView: (next: Viewport) => void,
    canZoom: () => boolean,
) {
    const zoom = useEffectEvent((event: WheelEvent, svg: SVGSVGElement) => {
        if (!canZoom() || event.deltaY === 0) return
        const anchor = screenToWorld(svg, event.clientX, event.clientY)
        if (!anchor) return
        event.preventDefault()
        setView(zoomView(view, event.deltaY < 0 ? 0.85 : 1 / 0.85, anchor))
    })
    useEffect(() => {
        const svg = ref.current
        if (!svg) return
        const handleWheel = (event: WheelEvent) => zoom(event, svg)
        svg.addEventListener('wheel', handleWheel, { passive: false })
        return () => svg.removeEventListener('wheel', handleWheel)
    }, [ref])
}
