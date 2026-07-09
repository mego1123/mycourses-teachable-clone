// Run multiple `ask` queries against the codewiki for jonradoff/lastsaas in sequence,
// threading history so each follow-up sees the prior Q&A.
import { CodeWikiClient } from '../repos/codewiki-mcp/dist/index.js'
import { writeFileSync } from 'node:fs'

const client = new CodeWikiClient({
  baseUrl: 'https://codewiki.google',
  timeoutMs: 120_000,
  maxRetries: 4,
  retryDelay: 600,
})

const REPO = 'jonradoff/lastsaas'

// Questions to walk the wiki from existing foundations toward a Teachable-style course platform.
const questions = [
  {
    id: 'q1_foundations',
    q: `What is the overall architecture of this repository? I want a clear picture of: (1) the tech stack (backend language/framework, frontend framework, database, message bus if any), (2) the top-level directory layout, (3) how the backend and frontend communicate (API style, auth), and (4) what the core domain abstractions are (tenants, users, memberships, plans, branding, billing, etc.). List the most important files/packages for each area.`,
  },
  {
    id: 'q2_multitenancy_branding',
    q: `How does multi-tenancy work end-to-end in this codebase? I specifically need to understand: (1) the Tenant model and how rows are scoped to a tenant (DB schema, middleware, repository layer), (2) how a request is associated with a tenant (subdomain? header? path? JWT claim?), (3) how tenant isolation is enforced in handlers and queries, (4) the white-label branding system — what is the Branding model, how is it loaded at runtime on the frontend (BrandingContext, BrandingThemeInjector), and how is it stored/served per-tenant on the backend. Cite concrete file paths.`,
  },
  {
    id: 'q3_billing_credits_stripe',
    q: `How do billing, plans, and credits work? I want: (1) the schema for Plan, Subscription, CreditBundle, UsageEvent, (2) how Stripe is integrated (which Stripe APIs are used — Checkout, Customer Portal, Webhooks?), (3) how credit-based metering works, (4) how a tenant changes plans or buys credits, (5) where Stripe webhook handlers live and what they update. Goal: understand what I would need to change to support a marketplace where a creator is paid for course sales (potentially via Stripe Connect) and the platform takes a commission.`,
  },
  {
    id: 'q4_auth_rbac_users',
    q: `How do authentication, users, and role-based access control work? Specifically: (1) the User, Membership, Invitation models and how a user belongs to a tenant with a role, (2) which roles exist out of the box and how RBAC is enforced in middleware, (3) how auth flows work (password, magic link, OAuth, MFA, sessions, JWT), (4) how the admin console differs from the tenant app, (5) impersonation. I want to understand what I'd need to add so that an end-learner has a different role/permission surface than a creator inside a tenant.`,
  },
  {
    id: 'q5_frontend_extensibility',
    q: `On the frontend, how are routes, pages, and contexts organized? I need: (1) the public vs. app vs. admin page split, (2) how TenantContext / BrandingContext / ThemeContext / AuthContext provide per-tenant UI, (3) how the BrandingThemeInjector applies tenant branding at runtime, (4) how the public landing page and CustomPage work (can a tenant already publish custom public pages?), (5) how API client attaches tenant/auth context to requests. Goal: figure out how I would add new "course storefront" public pages that are fully branded per-creator.`,
  },
  {
    id: 'q6_teachable_transformation',
    q: `I want to fork this codebase to build a multitenant course-selling SaaS like Teachable or Thinkific, where each "creator" gets their own branded course-selling website with custom domain, and learners sign up and buy courses. Given the existing tenant/branding/billing/auth foundations, give me a concrete mapping: which existing concepts should I reuse as-is, which should I rename or extend, and which new domain entities + DB tables + handlers + frontend pages do I need to add? Be specific about Course, Lesson, Section, Enrollment, CourseProgress, Coupon, Review, Payout, Creator/Student role split, custom domain routing, and Stripe Connect for creator payouts.`,
  },
  {
    id: 'q7_custom_domains',
    q: `How would I implement per-creator custom domains on top of this codebase? Today, how is the tenant resolved from an incoming HTTP request (middleware/tenant.go)? What changes are needed to: (1) add a CustomDomain table mapping custom domain -> tenant, (2) serve tenant branding and SSL/TLS for arbitrary custom domains, (3) handle the public storefront route vs. the platform admin/app routes, (4) handle preview/provisioning flows. Walk through the request lifecycle step-by-step with the specific files I'd touch.`,
  },
  {
    id: 'q8_course_content_storage',
    q: `How should I model and store course content (video lessons, PDFs, text) in this codebase? Today there is no media storage. Where would video uploads / file storage plug in (S3? a CDN?), how would I track playback progress per learner, gate access by purchase, and handle streaming/drip schedules? Reference any existing asset/binary/credit/usage patterns in the codebase that I can model after.`,
  },
]

async function run() {
  const history = []
  const answers = []
  for (const item of questions) {
    console.error(`\n=== ${item.id} ===`)
    console.error(item.q.slice(0, 200) + '...')
    try {
      const out = await client.askRepository(REPO, item.q, history)
      answers.push({ id: item.id, question: item.q, answer: out.data, meta: out.meta })
      history.push({ role: 'user', content: item.q })
      history.push({ role: 'assistant', content: out.data })
      console.error(`OK (${out.meta.totalBytes} bytes, ${out.meta.totalElapsedMs}ms)`)
    } catch (err) {
      answers.push({ id: item.id, question: item.q, error: err.message, cause: err.cause?.message })
      console.error(`FAIL: ${err.message}`)
    }
    // be polite to the RPC endpoint
    await new Promise(r => setTimeout(r, 1200))
  }
  writeFileSync('/home/z/my-project/download/lastsaas_wiki_qa.json', JSON.stringify(answers, null, 2))
  console.error(`\nWrote /home/z/my-project/download/lastsaas_wiki_qa.json`)
}

run().catch(err => { console.error(err); process.exit(1) })
