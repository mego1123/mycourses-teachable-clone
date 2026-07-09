// Run ONE ask query against the codewiki for jonradoff/lastsaas.
// Saves the answer to /home/z/my-project/download/wiki_answers/<id>.json
// Threads history from any prior answer files in the same dir, sorted by name prefix.
import { CodeWikiClient } from '../repos/codewiki-mcp/dist/index.js'
import { writeFileSync, readFileSync, readdirSync, mkdirSync } from 'node:fs'

const client = new CodeWikiClient({
  baseUrl: 'https://codewiki.google',
  timeoutMs: 150_000,
  maxRetries: 4,
  retryDelay: 800,
})

const REPO = 'jonradoff/lastsaas'
const OUTDIR = '/home/z/my-project/download/wiki_answers'
mkdirSync(OUTDIR, { recursive: true })

const id = process.argv[2]
const question = process.argv[3]

if (!id || !question) {
  console.error('Usage: node ask_one.mjs <id> "<question>"')
  process.exit(1)
}

// Build history from already-answered files (sorted by filename prefix NNN_id).
const existing = readdirSync(OUTDIR)
  .filter(f => f.endsWith('.json'))
  .sort()
const history = []
for (const f of existing) {
  try {
    const data = JSON.parse(readFileSync(`${OUTDIR}/${f}`, 'utf8'))
    if (data.question && data.answer) {
      history.push({ role: 'user', content: data.question })
      history.push({ role: 'assistant', content: data.answer })
    }
  } catch {}
}

console.error(`Asking ${id} (history items: ${history.length})...`)
const out = await client.askRepository(REPO, question, history)

const result = {
  id,
  question,
  answer: out.data,
  meta: out.meta,
  history_len: history.length,
}
writeFileSync(`${OUTDIR}/${id}.json`, JSON.stringify(result, null, 2))
console.error(`OK — ${out.meta.totalBytes} bytes in ${out.meta.totalElapsedMs}ms`)
console.log(out.data)
