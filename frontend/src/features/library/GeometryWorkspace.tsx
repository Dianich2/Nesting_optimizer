import { useEffect, useId, useState } from 'react'
import type { ReactNode } from 'react'

interface Props {
    toolbar: ReactNode
    tools: ReactNode
    properties: ReactNode
    status: ReactNode
    children: ReactNode
}

export function GeometryWorkspace({ toolbar, tools, properties, status, children }: Props) {
    const [toolsOpen, setToolsOpen] = useState(() => window.innerWidth > 700)
    const [propertiesOpen, setPropertiesOpen] = useState(() => window.innerWidth > 700)
    const toolsId = useId()
    const propertiesId = useId()

    useEffect(() => {
        const compact = window.matchMedia('(max-width: 700px)')
        const onChange = () => {
            if (compact.matches) {
                setToolsOpen(false)
                setPropertiesOpen(false)
            }
        }
        compact.addEventListener('change', onChange)
        return () => compact.removeEventListener('change', onChange)
    }, [])

    function toggle(side: 'tools' | 'properties') {
        const compact = window.matchMedia('(max-width: 700px)').matches
        if (side === 'tools') {
            setToolsOpen(!toolsOpen)
            if (compact && !toolsOpen) setPropertiesOpen(false)
        } else {
            setPropertiesOpen(!propertiesOpen)
            if (compact && !propertiesOpen) setToolsOpen(false)
        }
    }

    return (
        <div className="geometry-workspace">
            <header className="geometry-toolbar">
                {toolbar}
                <div className="geometry-panel-controls" role="group" aria-label="Боковые панели">
                    <button
                        type="button"
                        className="geometry-secondary"
                        aria-expanded={toolsOpen}
                        aria-controls={toolsId}
                        onClick={() => toggle('tools')}
                    >
                        Инструменты
                    </button>
                    <button
                        type="button"
                        className="geometry-secondary"
                        aria-expanded={propertiesOpen}
                        aria-controls={propertiesId}
                        onClick={() => toggle('properties')}
                    >
                        Свойства
                    </button>
                </div>
            </header>
            <div className="geometry-workspace-body">
                <aside
                    id={toolsId}
                    className="geometry-panel geometry-panel-tools"
                    aria-label="Инструменты геометрии"
                    hidden={!toolsOpen}
                >
                    {tools}
                </aside>
                <div className="geometry-canvas">{children}</div>
                <aside
                    id={propertiesId}
                    className="geometry-panel geometry-panel-properties"
                    aria-label="Свойства геометрии"
                    hidden={!propertiesOpen}
                >
                    {properties}
                </aside>
            </div>
            <footer className="geometry-status">{status}</footer>
        </div>
    )
}
