---
trigger: always_on
---

# Mod Manager (MM) — ContentSource adapter rules

## Package layout

Every external source lives under `backend/external/mm/<source>/`:

```
external/mm/<source>/
  adapter.go   -- implements mm.ContentSource
  dto.go       -- raw JSON structs from the external API
  mapper.go    -- dto -> mm.* conversion functions
  cache.go     -- optional, source-specific raw-response cache
```

The root `external/mm/` package is **read-only** — only domain types (`ContentSource`, `ContentType`, models, errors).

## Interface contract

An adapter **must** implement:

```go
func (a *Adapter) Name() string
func (a *Adapter) Capabilities() []mm.ContentType
func (a *Adapter) Search(ctx context.Context, req mm.SearchRequest) ([]mm.ContentSummary, int, error)
func (a *Adapter) GetContent(ctx context.Context, id string) (*mm.ContentDetail, error)
func (a *Adapter) GetVersions(ctx context.Context, contentID string, filter mm.VersionFilter) ([]mm.ContentVersion, error)
func (a *Adapter) GetVersion(ctx context.Context, versionID string) (*mm.ContentVersion, error)
func (a *Adapter) GetDownloadURL(ctx context.Context, versionID string) (string, error)
```

## ContentType mapping

The adapter decides which `ContentType` values it supports. Typical mapping for multi-type APIs:

- `mm.ContentMod` → `project_type:mod`
- `mm.ContentResourcepack` → `project_type:resourcepack`
- `mm.ContentShaderpack` → `project_type:shader`

If a source supports only mods, `Capabilities()` returns `[]mm.ContentType{mm.ContentMod}`.

## DTO → Domain mapping rules

- **Never** return raw DTOs from the adapter public methods.
- Map fields in `mapper.go` using plain functions (no methods on the adapter).
- Keep field names aligned with the external API; mapping layer handles normalization.
- Drop fields that have no equivalent in the domain model (e.g. source-specific metadata).

## Caching

- Cache **raw DTOs**, not mapped domain objects. Mapping is cheap; cache hit saves JSON parsing and HTTP round-trip.
- Use an in-memory TTL cache (see `external/mm/modrinth/cache.go` as reference).
- Cache key must include all request parameters that affect the response.

## Error handling

- Return `mm.ErrNotFound` when the upstream API responds with 404 or equivalent.
- Return `mm.ErrRateLimited` when hitting rate limits.
- Return `mm.ErrUnsupportedContentType` if `Search` is called with a `ContentType` not in `Capabilities()`.
- Wrap other errors with context: `fmt.Errorf("sourceName action: %w", err)`.

## HTTP client

- Set a timeout (e.g. 15s) and implement retry with exponential backoff for 429/5xx.
- Reuse a single `*http.Client` per adapter instance.
- Pass `ctx` into `http.NewRequestWithContext` so requests respect cancellation.

## HTTP routing

Backend exposes these generic endpoints (all under auth):

- `GET /api/mm/sources`
- `GET /api/mm/{source}/search?type=...&query=...`
- `GET /api/mm/{source}/content/{id}`
- `GET /api/mm/{source}/content/{id}/versions`
- `GET /api/mm/{source}/version/{id}`
- `GET /api/mm/{source}/version/{id}/download`

**Note:** `content/{id}` (not `/{id}`) is required to avoid Go `ServeMux` pattern conflicts with `version/{id}`.

## Registration

Add the adapter in `backend/cmd/main.go`:

```go
h.ContentSources = map[string]mm.ContentSource{
    "modrinth": modrinth.NewAdapter(os.Getenv("MODRINTH_CUSTOM_URL")),
    // "curseforge": curseforge.NewAdapter(...),
}
```

The map key becomes the `{source}` path segment.

## Frontend

- Front-end composables must call `/api/mm/{source}/...`, never the old `/api/modrinth/...` paths.
- The response shape is normalized; frontend should map `id` → internal `project_id` if legacy components expect it.
