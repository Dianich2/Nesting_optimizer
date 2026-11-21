export function parseQuantityDrafts(drafts: Record<number, string>) {
    const patterns: { project_pattern_id: number; quantity: number }[] = []
    let total = 0
    for (const [id, raw] of Object.entries(drafts)) {
        const text = raw.trim()
        if (!text) continue
        const value = Number(text)
        if (!/^\d+$/.test(text) || !Number.isSafeInteger(value) || value < 0 || value > 500)
            return {
                patterns: [],
                error: 'Количество: укажите целое число от 0 до 500. 0 — не выбирать лекало.',
            }
        if (!value) continue
        total += value
        if (total > 500)
            return {
                patterns: [],
                error: 'Суммарно за один запуск можно запросить не более 500 деталей.',
            }
        patterns.push({ project_pattern_id: Number(id), quantity: value })
    }
    return { patterns, error: null }
}
