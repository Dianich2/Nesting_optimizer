import type { LibraryGeometry, LibraryPoint } from './types'
import { insideRing, signedArea, validatePolygonDraft } from './geometryDraft'

const NUMBER = '[-+]?(?:\\d*\\.\\d+|\\d+\\.?\\d*)(?:[eE][-+]?\\d+)?'
const NUMBER_REGEX = new RegExp(NUMBER, 'g')
const TOKEN_REGEX = new RegExp(`[a-zA-Z]|${NUMBER}`, 'g')

function parseNumbers(input: string): number[] {
    const matches = [...input.matchAll(NUMBER_REGEX)]
    let last = 0
    const values: number[] = []
    for (const match of matches) {
        if (!/^[\s,]*$/.test(input.slice(last, match.index))) {
            throw new Error('В SVG обнаружены неподдерживаемые данные координат.')
        }
        values.push(Number(match[0]))
        last = (match.index ?? 0) + match[0].length
    }
    if (!/^[\s,]*$/.test(input.slice(last)) || values.some(v => !Number.isFinite(v))) {
        throw new Error('Некорректные координаты в SVG.')
    }
    return values
}

function normalizeRing(points: LibraryPoint[]): LibraryPoint[] {
    if (points.length > 1) {
        const first = points[0]
        const last = points[points.length - 1]
        if (Math.abs(first.x - last.x) < 1e-9 && Math.abs(first.y - last.y) < 1e-9) {
            return points.slice(0, -1)
        }
    }
    return points
}

/** SVG path commands M/m, L/l, H/h, V/v, Z/z. Curves/arcs are intentionally rejected. */
export function parseLinearPath(d: string): LibraryPoint[][] {
    const tokens = [...d.matchAll(TOKEN_REGEX)]
    let last = 0
    const values: string[] = []
    for (const match of tokens) {
        if (!/^[\s,]*$/.test(d.slice(last, match.index))) {
            throw new Error('Не удалось прочитать SVG path. Поддерживаются только прямые отрезки.')
        }
        values.push(match[0])
        last = (match.index ?? 0) + match[0].length
    }
    if (!/^[\s,]*$/.test(d.slice(last))) throw new Error('Некорректный SVG path.')

    const isCommand = (v: string) => /^[a-zA-Z]$/.test(v)
    let command = ''
    let cursor = 0
    let position: LibraryPoint = { x: 0, y: 0 }
    let start: LibraryPoint = { x: 0, y: 0 }
    let ring: LibraryPoint[] = []
    const rings: LibraryPoint[][] = []

    while (cursor < values.length) {
        if (isCommand(values[cursor])) {
            command = values[cursor++]
        }
        if (!command) throw new Error('SVG path должен начинаться с M или m.')
        const upper = command.toUpperCase()
        if (!'MLHVZ'.includes(upper)) {
            throw new Error(`SVG содержит команду ${command}. Пока поддерживаются только M, L, H, V, Z.`)
        }
        if (upper === 'Z') {
            if (ring.length < 3) throw new Error('Замкнутый SVG-контур должен иметь минимум 3 вершины.')
            rings.push(normalizeRing(ring))
            ring = []
            position = start
            command = ''
            continue
        }

        const count = upper === 'H' || upper === 'V' ? 1 : 2
        if (cursor + count > values.length ||
            values.slice(cursor, cursor + count).some(isCommand)) {
            throw new Error('В SVG path недостаточно координат для команды.')
        }
        const args = values.slice(cursor, cursor + count).map(Number)
        if (args.some(n => !Number.isFinite(n))) throw new Error('Некорректные числа SVG.')
        cursor += count
        const relative = command === command.toLowerCase()
        let next: LibraryPoint
        if (upper === 'H') next = { x: relative ? position.x + args[0] : args[0], y: position.y }
        else if (upper === 'V') next = { x: position.x, y: relative ? position.y + args[0] : args[0] }
        else next = {
            x: relative ? position.x + args[0] : args[0],
            y: relative ? position.y + args[1] : args[1],
        }
        if (upper === 'M') {
            if (ring.length > 0) throw new Error('Незамкнутый SVG-контур: добавьте команду Z.')
            start = next
            command = relative ? 'l' : 'L'
        } else if (ring.length === 0) {
            throw new Error('После Z новый контур должен начинаться с M.')
        }
        ring.push(next)
        position = next
    }
    if (ring.length > 0) throw new Error('Незамкнутый SVG-контур: добавьте команду Z.')
    return rings
}

function numberAttr(element: Element, name: string, fallback: number): number {
    const raw = element.getAttribute(name)
    if (raw === null) return fallback
    const number = Number(raw)
    if (!Number.isFinite(number)) throw new Error(`Некорректный SVG-атрибут ${name}.`)
    return number
}

function insideDefinitions(element: Element): boolean {
    for (let parent = element.parentElement; parent; parent = parent.parentElement) {
        if (['defs', 'clipPath', 'mask', 'symbol'].includes(parent.localName)) return true
    }
    return false
}

/** No SVG markup is ever mounted in the application DOM. Only numeric geometry is extracted. */
export function importSVGGeometry(source: string, allowHoles: boolean): LibraryGeometry {
    const document = new DOMParser().parseFromString(source, 'image/svg+xml')
    if (document.querySelector('parsererror') || document.documentElement.localName !== 'svg') {
        throw new Error('Файл не является корректным SVG.')
    }
    const root = document.documentElement
    const elements = Array.from(root.querySelectorAll('*')).filter(element => !insideDefinitions(element))
    if (elements.some(element => element.hasAttribute('transform')) || root.hasAttribute('transform')) {
        throw new Error('SVG с transform пока не поддерживается. Сохраните контуры без трансформаций.')
    }
    if (elements.some(e => ['circle', 'ellipse', 'line', 'polyline', 'text', 'image', 'use'].includes(e.localName))) {
        throw new Error('SVG содержит неподдерживаемые фигуры. Нужны polygon, rect или path из прямых линий.')
    }
    const shapes = elements.filter(e => ['polygon', 'rect', 'path'].includes(e.localName))
    if (shapes.length === 0) throw new Error('В SVG не найдено ни одного подходящего контура.')

    const rings: LibraryPoint[][] = []
    for (const shape of shapes) {
        if (shape.localName === 'path') {
            rings.push(...parseLinearPath(shape.getAttribute('d') ?? ''))
        } else if (shape.localName === 'polygon') {
            const coords = parseNumbers(shape.getAttribute('points') ?? '')
            if (coords.length % 2 !== 0) throw new Error('SVG polygon содержит нечётное число координат.')
            const points: LibraryPoint[] = []
            for (let i = 0; i < coords.length; i += 2) {
                points.push({ x: coords[i], y: coords[i + 1] })
            }
            rings.push(normalizeRing(points))
        } else {
            if (numberAttr(shape, 'rx', 0) > 0 || numberAttr(shape, 'ry', 0) > 0) {
                throw new Error('Скруглённые SVG-прямоугольники пока не поддерживаются.')
            }
            const x = numberAttr(shape, 'x', 0)
            const y = numberAttr(shape, 'y', 0)
            const width = numberAttr(shape, 'width', 0)
            const height = numberAttr(shape, 'height', 0)
            if (width <= 0 || height <= 0) throw new Error('Размеры SVG-прямоугольника должны быть положительными.')
            rings.push([
                { x, y }, { x: x + width, y },
                { x: x + width, y: y + height }, { x, y: y + height },
            ])
        }
    }
    if (rings.length === 0) throw new Error('В SVG отсутствуют замкнутые контуры.')
    const sorted = [...rings].sort((a, b) => Math.abs(signedArea(b)) - Math.abs(signedArea(a)))
    const exterior = sorted[0]
    const holes = sorted.slice(1)
    if (holes.some(hole => !hole.every(p => insideRing(p, exterior)))) {
        throw new Error('SVG содержит несколько отдельных фигур. Нужен один внешний контур с внутренними отверстиями.')
    }
    const geometry: LibraryGeometry = {
        exterior: { points: exterior },
        holes: holes.map(points => ({ points })),
    }
    const problem = validatePolygonDraft(geometry, allowHoles)
    if (problem) throw new Error(problem)
    return geometry
}
