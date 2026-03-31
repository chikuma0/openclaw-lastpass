# Secret Structure Guide

`openclaw-lastpass` works best when LastPass entries are structured for machine use on purpose. The goal is not to mirror every possible LastPass object shape. The goal is to make machine secrets easy to name, review, and resolve without ambiguity.

## Default Recommendation

Use one of these two patterns:

1. one LastPass entry for one secret, stored in the built-in `password` field
2. one LastPass entry for one repo x environment, with several related custom fields

Pick the smallest pattern that stays understandable six months later.

## Pattern 1: One Secret Per Entry

Use this when the entry exists only to hold a single API key, token, or password.

Example LastPass entry:

- name: `Machine Secrets/acme/prod/OPENAI_API_KEY`
- password field: the actual OpenAI API key

Example resolver mapping:

```json
{
  "providers/openai/apiKey": {
    "entry": "Machine Secrets/acme/prod/OPENAI_API_KEY",
    "field": "password"
  }
}
```

This is the best choice when:

- you only need one value from the item
- the value rotates independently from everything else
- you want the clearest possible mapping

## Pattern 2: One Entry Per Repo x Environment

Use this when several related machine secrets belong to the same runtime boundary.

Example LastPass entry:

- name: `Machine Secrets/acme/prod/runtime`
- custom fields:
  - `ANTHROPIC_API_KEY`
  - `DATABASE_URL`
  - `SUPABASE_SERVICE_ROLE`

Example resolver mapping:

```json
{
  "providers/anthropic/apiKey": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "ANTHROPIC_API_KEY"
  },
  "services/acme/databaseUrl": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "DATABASE_URL"
  },
  "supabase/serviceRole": {
    "entry": "Machine Secrets/acme/prod/runtime",
    "field": "SUPABASE_SERVICE_ROLE"
  }
}
```

This is the best choice when:

- one app or service needs several related machine secrets
- the values are naturally reviewed together
- you want fewer LastPass items without losing clarity

## Field Guidance

Built-in field names:

- `password`
- `username`
- `notes`

Anything else is treated as a custom field name.

Recommendations:

- Prefer `password` for single opaque secrets.
- Prefer custom fields for structured runtime values such as `DATABASE_URL`, `SERVICE_URL`, or `SUPABASE_SERVICE_ROLE`.
- Use `notes` only when you are intentionally storing a value in a secure note body or cleaning up older LastPass data.
- Prefer lowercase built-in names in mapping files so the intent is obvious at a glance.

## Naming Guidance

Good entry naming reduces ambiguity and makes discovery more useful.

Prefer names and paths that include:

- the repo or system name
- the environment
- the purpose of the secret

Examples:

- `Machine Secrets/acme/prod/OPENAI_API_KEY`
- `Machine Secrets/acme/staging/runtime`
- `Machine Secrets/shared/github_pat_repo_admin`

Avoid names like:

- `OpenAI`
- `token`
- `prod secret`
- `database`

If a mapping would still be ambiguous, use the LastPass item ID instead of the entry name.

## Discovery And Apply Caveat

The final resolver mapping can point multiple secret IDs at the same LastPass entry. That works well for bundled custom-field entries.

Today, `discover` and `apply` stay conservative and only draft one suggestion per LastPass item. That means a bundled entry may still require a small hand edit to the final `mapping.json` for additional fields after `apply` writes the first approved mapping.

That tradeoff is intentional. The project prefers a smaller, safer draft schema over trying to guess several mappings from one vault item.

## Recommended Review Habits

- Keep entry names stable.
- Keep custom field names stable.
- Keep one human in the loop before `apply`.
- Run `openclaw-lastpass doctor` after major mapping changes.
- Map only the secret IDs your CLI or OpenClaw workflow actually needs.

Related examples:

- [`examples/mapping.example.json`](../examples/mapping.example.json)
- [`examples/mapping.final.example.json`](../examples/mapping.final.example.json)
- [`examples/mapping.draft.example.json`](../examples/mapping.draft.example.json)
