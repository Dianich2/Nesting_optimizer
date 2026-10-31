
import type { ReactNode } from 'react'
import { Link } from 'react-router'

import './BackLink.css'

interface BackLinkProps {
    to: string
    children?: ReactNode
}

export function BackLink({
    to,
    children = 'Назад',
}: BackLinkProps) {
    return (
        <Link to={to} className="back-link">
            <span aria-hidden="true">←</span>
            <span>{children}</span>
        </Link>
    )
}
