import { CodeWikiClient } from '../repos/codewiki-mcp/dist/index.js'
import { writeFileSync, mkdirSync } from 'node:fs'

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
  console.error('Usage: node ask_one_clean.mjs <id> "<question>"')
  process.exit(1)
}

// NO history at all - fresh conversation each time
console.error(`Asking ${id} (no history)...`)
const out = await client.askRepository(REPO, question, [])

const result = { id, question, answer: out.data, meta: out.meta, history_len: 0 }
writeFileSync(`${OUTDIR}/${id}.json`, JSON.stringify(result, null, 2))
console.error(`OK — ${out.meta.totalBytes} bytes in ${out.meta.totalElapsedMs}ms`)
console.log(out.data)
