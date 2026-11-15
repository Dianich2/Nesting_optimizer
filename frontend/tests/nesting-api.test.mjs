import assert from 'node:assert/strict'
import test from 'node:test'
import { allProjectPatterns } from '../src/features/nesting/api.ts'

test('loads every page using the selected project ID', async () => {
    const paths = []
    const request = async (path) => {
        paths.push(path)
        return { items: [{ id: paths.length, project_id: 7 }], total_pages: 3 }
    }
    const patterns = await allProjectPatterns(request, 7)
    assert.deepEqual(patterns.map((item) => item.id), [1, 2, 3])
    assert.deepEqual(paths, [1, 2, 3].map((page) => `/projects/7/patterns?page=${page}&page_size=100`))
})

test('pagination errors propagate without returning a partial library', async () => {
    let calls = 0
    const error = new Error('request failed')
    const request = async () => {
        if (++calls === 2) throw error
        return { items: [{ id: 1 }], total_pages: 2 }
    }
    await assert.rejects(allProjectPatterns(request, 9), error)
})
