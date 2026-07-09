// Thin CLI wrapper around codewiki-mcp's CodeWikiClient to query any repo wiki.
// Usage:
//   node wiki_query.mjs search "<query>"
//   node wiki_query.mjs fetch "<owner/repo|url>" [aggregate|pages]
//   node wiki_query.mjs ask "<owner/repo|url>" "<question>" [--history path/to/history.json]
//
// Output: pretty-printed JSON to stdout, raw answer text on stdout for `ask` when --plain.

import { CodeWikiClient } from '../repos/codewiki-mcp/dist/index.js'
import { writeFileSync, readFileSync } from 'node:fs'

const client = new CodeWikiClient({
  baseUrl: 'https://codewiki.google',
  timeoutMs: 90_000,
  maxRetries: 4,
  retryDelay: 500,
})

const [mode, ...rest] = process.argv.slice(2)

function parseFlags(args) {
  const positional = []
  const flags = {}
  for (let i = 0; i < args.length; i++) {
    const a = args[i]
    if (a.startsWith('--')) {
      const key = a.slice(2)
      const val = args[i + 1] && !args[i + 1].startsWith('--') ? args[i + 1] : true
      flags[key] = val
      if (val !== true) i++
    } else {
      positional.push(a)
    }
  }
  return { positional, flags }
}

async function main() {
  if (mode === 'search') {
    const { positional, flags } = parseFlags(rest)
    const query = positional[0]
    const limit = flags.limit ? Number(flags.limit) : 10
    if (!query) throw new Error('search needs a query')
    const out = await client.searchRepositories(query, limit)
    console.log(JSON.stringify(out, null, 2))
    return
  }

  if (mode === 'fetch') {
    const { positional } = parseFlags(rest)
    const repo = positional[0]
    const outMode = positional[1] || 'aggregate'
    if (!repo) throw new Error('fetch needs a repo')
    const result = await client.fetchRepository(repo)
    if (outMode === 'pages') {
      console.log(JSON.stringify({
        repo: result.data.repo,
        commit: result.data.commit,
        canonicalUrl: result.data.canonicalUrl,
        sections: result.data.sections.map(s => ({
          title: s.title,
          level: s.level,
          anchor: s.anchor,
          summary: s.summary,
          markdown: s.markdown,
          diagramCount: s.diagramCount,
        })),
        meta: result.meta,
      }, null, 2))
    } else {
      // aggregate: stitch together into one markdown document
      const md = [
        `# Wiki: ${result.data.repo}`,
        ``,
        `> commit: ${result.data.commit ?? 'unknown'} | canonical: ${result.data.canonicalUrl ?? 'n/a'}`,
        ``,
        ...result.data.sections.map(s => `## ${s.title}\n\n${s.markdown}`),
      ].join('\n')
      console.log(md)
    }
    return
  }

  if (mode === 'ask') {
    const { positional, flags } = parseFlags(rest)
    const repo = positional[0]
    const question = positional[1]
    if (!repo || !question) throw new Error('ask needs <repo> <question>')
    let history = []
    if (flags.history) {
      history = JSON.parse(readFileSync(flags.history, 'utf8'))
    }
    const out = await client.askRepository(repo, question, history)
    if (flags.plain) {
      console.log(out.data)
    } else {
      console.log(JSON.stringify(out, null, 2))
    }
    return
  }

  console.error('Usage: node wiki_query.mjs [search|fetch|ask] ...')
  process.exit(1)
}

main().catch(err => {
  console.error('Error:', err.message || err)
  if (err.cause) console.error('Cause:', err.cause)
  process.exit(1)
})
