// Run ONE ask query against the codewiki for jonradoff/lastsaas.
// Only threads the LAST 2 Q&A pairs as context (to avoid hitting RPC safety limits).
import { CodeWikiClient } from '../repos/codewiki-mcp/dist/index.js'
import { writeFileSync, readFileSync, readdirSync, mkdirSync } from 'node:fs'

const client = new CodeWikiClient({
  baseUrl: 'https://codewiki.google',
  timeoutMs: 180_000,
  maxRetries: 4,
  retryDelay: 1000,
})

const REPO = 'jonradoff/lastsaas'
const OUTDIR = '/home/z/my-project/download/wiki_answers'
mkdirSync(OUTDIR, { recursive: true })

const id = process.argv[2]
const question = process.argv[3]

if (!id || !question) {
  console.error('Usage: node ask_one_short.mjs <id> "<question>"')
  process.exit(1)
}

// Build history only from the LAST 2 answer files (to stay under RPC limits)
const existing = readdirSync(OUTDIR)
  .filter(f => f.endsWith('.json'))
  .sort()
  .slice(-2)
const history = []
for (const f of existing) {
  try {
    const data = JSON.parse(readFileSync(`${OUTDIR}/${f}`, 'utf8'))
    if (data.question && data.answer && data.answer.length > 200) {
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
