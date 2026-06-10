# terraform-provider-ngrok v2 — Design Document

## 1. Executive Summary

Rebuild the ngrok Terraform provider from scratch to support the modern ngrok platform (cloud endpoints, traffic policy, vaults/secrets, AI gateway) using the Terraform Plugin Framework, hand-written resources following consistent patterns, and the official `ngrok-api-go` client library.

---

## 2. Problems with the Existing Provider

| Issue | Detail |
|---|---|
| **Outdated Terraform SDK** | Uses `terraform-plugin-sdk/v2` (SDKv2 v2.6.1) — HashiCorp recommends the Plugin Framework for all new providers. SDKv2 is in maintenance mode. |
| **Custom REST client** | Hand-rolled `restapi/` package instead of using the official `ngrok-api-go` library (now at v9). Duplicates types, methods, and auth logic. |
| **Code-generated but frozen** | Every file says `// Code generated for API Clients. DO NOT EDIT.` but there's no generator in the repo — code was generated once externally and committed. No way to re-generate when APIs change. |
| **Missing major resources** | No support for: Cloud Endpoints, Traffic Policy, Vaults, Secrets, Kubernetes Operators. Still exposes deprecated edges/endpoint configurations. |
| **Stale Go version** | `go 1.18` — current stable is 1.23+. |
| **No acceptance tests** | No `_test.go` files in the repo. |
| **No import support** | Resources don't implement `Importer`, so `terraform import` doesn't work. |
| **No data sources** | Zero data sources — users can't reference existing resources. |

---

## 3. Architecture

### 3.1 High-Level Overview

```
┌──────────────────────────────────────────────────────────┐
│              terraform-provider-ngrok v2                  │
│                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │  Provider    │  │  Resources   │  │  Data Sources  │  │
│  │  (config,    │  │  (hand-      │  │  (read-only    │  │
│  │   auth)      │  │   written)   │  │   lookups)     │  │
│  └──────┬───────┘  └──────┬───────┘  └───────┬────────┘  │
│         │                 │                   │           │
│         │    ┌────────────┴──────────────┐    │           │
│         │    │  Shared Helpers            │    │           │
│         │    │  (expand/flatten, errors,  │    │           │
│         │    │   validators, plan mods)   │    │           │
│         │    └────────────┬──────────────┘    │           │
│         │                 │                   │           │
│         └─────────┬───────┴───────────────────┘           │
│                   ▼                                       │
│         ┌──────────────────┐                              │
│         │  ngrok-api-go/v9 │  (official Go API client)    │
│         └──────────────────┘                              │
│                   │                                       │
│  ┌────────────────┴────────────────────────────────────┐  │
│  │  Terraform Plugin Framework (hashicorp/terraform-   │  │
│  │  plugin-framework)                                  │  │
│  └─────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

### 3.2 Key Technology Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Terraform SDK | **Plugin Framework** (`terraform-plugin-framework`) | HashiCorp's recommended path. Better type safety, native support for nested objects, plan modifiers, validators. |
| API Client | **`ngrok-api-go/v9`** | Official, maintained, code-generated from the same OpenAPI spec. Eliminates the custom `restapi/` package. |
| Resource authoring | **Hand-written with shared helpers** | ~20 resources is well within hand-written range. `ngrok-api-go` already provides typed structs/methods (itself code-gen'd from the OpenAPI spec), so resources are straightforward CRUD wiring. Consistent patterns enforced via shared helpers and code review, not a generator. |
| Go Version | **1.23+** | Current stable, required by modern dependencies. |
| Testing | **Acceptance tests + unit tests** | Use `terraform-plugin-testing` for acceptance tests against the real API. Unit tests for schema logic. |
| Docs | **`tfplugindocs`** | Auto-generate registry docs from schema descriptions and examples. |
| CI | **GitHub Actions** | Lint, test, goreleaser for releases. |

---

## 4. Resource & Data Source Coverage

### 4.1 Full Resource Matrix

Resources derived from the ngrok OpenAPI spec. **Bold** = new (not in v1). Priority tiers determine implementation order.

#### Tier 1 — Core (MVP)
| Resource | Data Source | Notes |
|---|---|---|
| `ngrok_reserved_domain` | `ngrok_reserved_domain` | Carries over from v1 |
| `ngrok_reserved_addr` | `ngrok_reserved_addr` | Carries over from v1 |
| **`ngrok_cloud_endpoint`** | **`ngrok_cloud_endpoint`** | Cloud endpoints with traffic_policy. Key new resource. |
| `ngrok_certificate_authority` | `ngrok_certificate_authority` | Carries over |
| `ngrok_tls_certificate` | `ngrok_tls_certificate` | Carries over |
| `ngrok_api_key` | `ngrok_api_key` | Carries over |
| `ngrok_credential` | `ngrok_credential` | Carries over (tunnel auth tokens) |
| `ngrok_ip_policy` | `ngrok_ip_policy` | Carries over |
| `ngrok_ip_policy_rule` | `ngrok_ip_policy_rule` | Carries over |
| `ngrok_ip_restriction` | `ngrok_ip_restriction` | Carries over |
| `ngrok_service_user` | `ngrok_service_user` | Carries over (API resource: bot_users) |
| `ngrok_agent_ingress` | `ngrok_agent_ingress` | Carries over |
| `ngrok_event_destination` | `ngrok_event_destination` | Carries over |
| `ngrok_event_subscription` | `ngrok_event_subscription` | Carries over |

#### Tier 2 — Platform Features
| Resource | Data Source | Notes |
|---|---|---|
| **`ngrok_vault`** | **`ngrok_vault`** | New — secret management |
| **`ngrok_secret`** | **`ngrok_secret`** | New — secret management |
| `ngrok_ssh_certificate_authority` | `ngrok_ssh_certificate_authority` | Carries over |
| `ngrok_ssh_credential` | `ngrok_ssh_credential` | Carries over |
| `ngrok_ssh_host_certificate` | `ngrok_ssh_host_certificate` | Carries over |
| `ngrok_ssh_user_certificate` | `ngrok_ssh_user_certificate` | Carries over |

#### Tier 3 — Observability & Advanced
| Resource | Data Source | Notes |
|---|---|---|
| — | **`ngrok_tunnels`** | Read-only — list active tunnels |
| — | **`ngrok_tunnel_sessions`** | Read-only — list active sessions |
| — | **`ngrok_application_sessions`** | Read-only |
| — | **`ngrok_application_users`** | Read-only |
| **`ngrok_kubernetes_operator`** | **`ngrok_kubernetes_operator`** | New |

> **Note:** Edges (HTTPS/TCP/TLS), endpoint configurations, and backends (failover, weighted, tunnel group, HTTP response) are **deprecated** by the ngrok platform in favor of cloud endpoints + traffic policy. They are intentionally excluded from v2.

---

## 5. Resource Development Approach

### 5.1 Why Hand-Written (Not Code-Generated)

The existing provider was code-generated externally and committed as frozen output — when the API evolved, the provider couldn't be updated because the generator wasn't in the repo. Code generation also struggled to express Terraform-specific concerns (ForceNew, plan modifiers, custom validators, nested object semantics).

Premier Terraform providers (AWS ~1,500 resources, Azure ~800, Cloudflare ~150) are hand-written. The pattern is clear:

1. Use the official typed SDK client — **`ngrok-api-go/v9`** already provides typed structs and methods for every API resource, generated from the OpenAPI spec. The API layer is already abstracted.
2. Hand-write resources following strict internal conventions — each resource is ~150-250 lines of straightforward CRUD wiring.
3. Invest in shared helpers — reusable expand/flatten functions, error handling, validators, and plan modifiers eliminate boilerplate without the overhead of maintaining a generator.

With ~20 resources, a custom code generator would cost more to build and maintain than it saves.

### 5.2 Shared Helpers

Invest in a `helpers.go` package that handles repeated patterns across resources:

| Helper | Purpose |
|---|---|
| `expandRef` / `flattenRef` | Convert between `ngrok.Ref{ID: "..."}` and `types.String` for API references |
| `flattenTimestamp` | Consistent `created_at` handling |
| `handleNotFoundError` | On Read, detect 404 → remove from state (standard Terraform pattern) |
| `setID` | Set resource ID from API response |
| `stringPtrFromFramework` / `stringFromPtr` | Convert between Framework `types.String` and `*string` for optional API fields |

**Null vs empty string handling:** In the Plugin Framework, `types.String` has three states: **null** (not set), **unknown** (computed, not yet known), and **value** (set, including empty `""`). This distinction is critical for API calls:
- Null → don't send the field (omit from request) → API keeps its default or existing value
- Empty `""` → send empty string → API clears the field
- Value → send the value

`stringPtrFromFramework` returns `nil` for null/unknown (field omitted from API request), `*""` for empty, and `*"value"` for set. Getting this wrong causes silent bugs: sending empty strings when the user didn't set a field, which clears API defaults.

### 5.3 Resource Conventions

Every resource follows the same file structure and conventions:

```go
// internal/provider/resource_reserved_domain.go

// 1. Type definition with tfsdk model
type reservedDomainResourceModel struct { ... }

// 2. Resource type implementing framework interfaces
type reservedDomainResource struct {
    client *ngrok.Client
}

// 3. Metadata — returns "ngrok_reserved_domain"
// 4. Schema — attributes, descriptions, plan modifiers, validators
// 5. Create — expand model → API call → flatten response → set state
// 6. Read — API call → handle 404 → flatten → set state
// 7. Update — expand changed fields → API call → flatten → set state
// 8. Delete — API call → handle 404
// 9. ImportState — passthrough ID
```

### 5.4 Keeping Up with API Changes

- **`ngrok-api-go/v9`** is the sync point — when ngrok adds new API resources or fields, they release a new version of the Go client. We bump the dependency and add/update resources as needed.
- The OpenAPI spec at `github.com/ngrok/ngrok-openapi` serves as a **reference** for understanding API shapes, not as a code generation input.
- New resources are added by copying an existing resource file and adapting it — with consistent patterns this takes 30-60 minutes per resource.

---

## 6. Project Structure

```
terraform-provider-ngrok/
├── internal/
│   └── provider/
│       ├── provider.go              # Provider definition, config, auth
│       ├── provider_test.go
│       ├── helpers.go               # Shared expand/flatten, error handling
│       ├── validators.go            # Custom validators (e.g., JSON syntax)
│       ├── resource_reserved_domain.go
│       ├── resource_reserved_domain_test.go
│       ├── datasource_reserved_domain.go
│       ├── resource_cloud_endpoint.go     # Cloud endpoints + traffic policy
│       ├── resource_cloud_endpoint_test.go
│       └── ... (one file per resource/datasource)
├── templates/                       # tfplugindocs templates
│   ├── index.md.tmpl
│   └── resources/
│       └── reserved_domain.md.tmpl
├── examples/
│   └── resources/
│       ├── ngrok_reserved_domain/
│       │   └── resource.tf
│       ├── ngrok_cloud_endpoint/
│       │   └── resource.tf
│       └── ...
├── docs/                            # Generated by tfplugindocs
├── .github/
│   └── workflows/
│       ├── ci.yml                   # lint + unit tests
│       ├── acceptance.yml           # acceptance tests (needs API key)
│       └── release.yml              # goreleaser
├── .goreleaser.yml
├── go.mod
├── go.sum
├── main.go                          # Entry point
├── Makefile
└── DESIGN.md                        # This file
```

---

## 7. Implementation Plan

### Phase 1: Foundation (Week 1-2)
- [ ] Initialize Go module with `go 1.23`, Plugin Framework, `ngrok-api-go/v9`
- [ ] Implement provider configuration (api_key, api_base_url)
- [ ] Build shared helpers (`helpers.go`, `validators.go`)
- [ ] Build the first resource: `ngrok_reserved_domain` (resource + data source + import + acceptance test) — this becomes the reference pattern for all other resources
- [ ] Set up CI (lint, unit test, acceptance test matrix)
- [ ] Set up `tfplugindocs` generation
- [ ] Set up goreleaser for publishing

### Phase 2: Tier 1 Resources (Week 2-4)
- [ ] `ngrok_cloud_endpoint` (cloud endpoints with traffic_policy support)
- [ ] `ngrok_reserved_addr`
- [ ] `ngrok_api_key`, `ngrok_credential`
- [ ] `ngrok_certificate_authority`, `ngrok_tls_certificate`
- [ ] `ngrok_ip_policy`, `ngrok_ip_policy_rule`, `ngrok_ip_restriction`
- [ ] `ngrok_service_user`, `ngrok_agent_ingress`
- [ ] `ngrok_event_destination`, `ngrok_event_subscription`
- [ ] Acceptance tests for all Tier 1 resources
- [ ] **Alpha release** to Terraform Registry

### Phase 3: Tier 2 Resources (Week 4-5)
- [ ] `ngrok_vault`, `ngrok_secret`
- [ ] SSH resources (`ngrok_ssh_certificate_authority`, `ngrok_ssh_credential`, `ngrok_ssh_host_certificate`, `ngrok_ssh_user_certificate`)
- [ ] Acceptance tests for all Tier 2

### Phase 4: Tier 3 & Polish (Week 5-6)
- [ ] Read-only data sources (tunnels, sessions, application users)
- [ ] `ngrok_kubernetes_operator`
- [ ] Migration guide from v1 → v2
- [ ] Full documentation review
- [ ] **GA release**

---

## 8. Key Design Details

### 8.1 Provider Configuration

```go
type ngrokProviderModel struct {
    APIKey     types.String `tfsdk:"api_key"`
    APIBaseURL types.String `tfsdk:"api_base_url"`
}
```

- `api_key`: Required. Sourced from `NGROK_API_KEY` env var.
- `api_base_url`: Optional. Defaults to `https://api.ngrok.com`. Sourced from `NGROK_API_BASE_URL`.

The provider creates an `ngrok-api-go` `ClientConfig` and passes it through the framework's provider data mechanism.

**User-Agent:** The provider sets a custom User-Agent header on all API requests: `terraform-provider-ngrok/1.0.0 terraform/<version>`. This lets ngrok distinguish Terraform traffic from other API clients for analytics and debugging. The `ngrok-api-go` client supports custom HTTP headers via the `ClientConfig`.

**Minimum Terraform version:** Terraform ≥ 1.1 (required for `moved` blocks used in migration). Specified in the provider's `terraform` block and enforced by the Plugin Framework.

### 8.2 Cloud Endpoints + Traffic Policy

The `ngrok_cloud_endpoint` resource is the most important new addition. `traffic_policy` is a JSON string field, authored with Terraform's built-in `jsonencode()`.

```hcl
resource "ngrok_cloud_endpoint" "example" {
  url = "https://app.example.com"

  traffic_policy = jsonencode({
    on_http_request = [
      {
        actions = [
          {
            type   = "rate-limit"
            config = {
              name      = "global"
              algorithm = "sliding_window"
              capacity  = 100
              rate      = "60s"
            }
          }
        ]
      }
    ]
  })

  bindings       = ["public"]
  description    = "Production API endpoint"
  metadata       = jsonencode({ team = "platform" })
}
```

**Key schema decisions:**

| Attribute | Type | Notes |
|---|---|---|
| `url` | Required, **ForceNew** | Changing the URL orphans the old auto-created domain and creates a new one. ForceNew makes this explicit in the plan. |
| `traffic_policy` | Required | The API requires a traffic policy for cloud endpoints. |
| `bindings` | Optional, default `["public"]` | Controls whether the endpoint is `"public"` (internet-facing) or `"internal"` (accessible only within your ngrok network). |
| `pooling_enabled` | Optional, default `false` | Whether the endpoint allows connection pooling across multiple agents. |
| `description` | Optional | Human-readable description. |
| `metadata` | Optional | Arbitrary user-defined machine-readable data (max 4096 bytes). |
| `domain_id` | Computed | ID of the associated reserved domain (auto-created or pre-existing). |

**Why a string field (not structured HCL blocks):**

Traffic policy is a rich, evolving DSL with 20+ action types, each with a different config shape. Modeling it as nested HCL blocks would tightly couple the provider to the policy schema (every new action type = provider release), prevent policy sharing across endpoints via locals/variables, and risk data loss on API round-trips. The string approach is the industry standard — AWS IAM policies, Azure ARM templates, and Cloudflare Workers all use the same pattern.

**Verified: the ngrok API returns `traffic_policy` byte-for-byte identical to what was sent.** No key reordering, no injected defaults, no normalization. This eliminates the phantom diff problem entirely — no semantic JSON comparison or normalization logic needed.

**UX with `jsonencode()`:**
- Terraform diffs the HCL expression, showing exactly which fields changed (not an opaque string blob)
- TF resource interpolation works: `ip_policies = [ngrok_ip_policy.office.id]`
- Policies can be shared across endpoints via `locals {}` or Terraform modules
- `jsonencode()` output is deterministic (sorted keys, consistent whitespace)

**Validation:** A custom plan-time validator checks JSON syntax. Structural validation (valid action types, required fields like `algorithm`) is left to the API — it returns clear error messages (e.g., `"algorithm must be sliding_window, was ."`).

**Documentation should recommend `jsonencode()` over `file()`.** The `file()` path works but produces opaque string diffs. `templatefile()` is an acceptable alternative when policies are managed as separate files.

### 8.3 Endpoint + Domain Lifecycle

Creating an endpoint via the ngrok API **auto-creates a reserved domain** as a side effect. Deleting the endpoint **does not delete the domain** — it persists. This is API behavior, not something the provider controls.

Each resource manages exactly one thing:
- `ngrok_cloud_endpoint` → creates/deletes endpoints only
- `ngrok_reserved_domain` → creates/deletes domains only

The provider does not attempt cross-resource cleanup. If the API auto-creates a domain as a side effect, Terraform doesn't track or delete it. Both usage patterns are supported:

**Simple (endpoint only):**
```hcl
resource "ngrok_cloud_endpoint" "app" {
  url            = "https://myapp.ngrok.app"
  traffic_policy = jsonencode({ ... })
}
# Domain auto-created by API, not managed by Terraform.
# terraform destroy deletes the endpoint but the domain persists.
```

**Full lifecycle control (recommended for production):**
```hcl
resource "ngrok_reserved_domain" "app" {
  domain = "app.example.com"
}

resource "ngrok_cloud_endpoint" "app" {
  url            = "https://${ngrok_reserved_domain.app.domain}"
  traffic_policy = jsonencode({ ... })
}
# terraform destroy deletes the endpoint first, then the domain.
```

The endpoint resource exposes `domain_id` as a computed attribute so users can see the associated domain regardless of which pattern they use.

### 8.4 Data Sources

Data sources support lookup by **ID** (direct GET) or by **filterable field** (server-side CEL filtering). The ngrok API supports CEL filter expressions on list endpoints (`?filter=obj.domain == "..."`), so lookups are efficient — no client-side pagination needed.

```hcl
# Lookup by ID
data "ngrok_reserved_domain" "by_id" {
  id = "rd_2example"
}

# Lookup by domain name
data "ngrok_reserved_domain" "by_name" {
  domain = "app.example.com"
}

# Lookup endpoint by URL
data "ngrok_cloud_endpoint" "by_url" {
  url = "https://app.example.com"
}
```

**Implementation:** each data source accepts `id` (optional) or a resource-specific lookup field (optional). If `id` is set, call GET directly. Otherwise, call List with a CEL filter, expect exactly one result, error if zero or multiple matches.

**Filterable fields per resource** (verified against the API):

| Data Source | Lookup fields (besides `id`) |
|---|---|
| `ngrok_reserved_domain` | `domain` |
| `ngrok_reserved_addr` | `addr` |
| `ngrok_cloud_endpoint` | `url`, `name` |
| `ngrok_ip_policy` | `description` |
| `ngrok_agent_ingress` | `domain` |
| `ngrok_vault` | `name` |
| `ngrok_secret` | `name` |
| `ngrok_tls_certificate` | `subject_common_name` |

### 8.5 Import Support

Every resource implements `ImportState` using the resource ID:

```bash
terraform import ngrok_reserved_domain.example rd_2example
```

### 8.6 Testing Strategy

| Test Type | Tool | Scope |
|---|---|---|
| Unit tests | `go test` | Schema validation, expand/flatten logic |
| Acceptance tests | `terraform-plugin-testing` | Full CRUD lifecycle against ngrok API |
| Linting | `golangci-lint` | Code quality |
| Doc validation | `tfplugindocs validate` | Registry docs |

Acceptance tests require `NGROK_API_KEY` env var and run in CI with a dedicated test account.

---

## 9. Migration from v0.x

### 9.1 Versioning Strategy

The current provider is v0.7.0. We release the rebuild as **v1.0.0** under the same registry namespace (`ngrok/ngrok`). Semantic versioning protects existing users:

- Users with `version = "~> 0.7"` → Terraform will **not** auto-upgrade to v1.0.0
- Users with no version constraint → will get v1.0.0 on next `terraform init`, but `terraform plan` will surface all issues before any changes are applied
- Users opt in explicitly by changing their version constraint to `version = "~> 1.0"`

This is the standard approach across the Terraform ecosystem (AWS v4→v5, Azure v3→v4, Cloudflare v4→v5).

### 9.2 Breaking Changes

| Change | What the user does |
|---|---|
| Removed resources (edges, backends, endpoint configurations) | `terraform state rm ngrok_edge_https.example`, delete the HCL block, replace with `ngrok_cloud_endpoint` + traffic policy |
| Removed fields (`region`, `http_endpoint_configuration_id`, `https_endpoint_configuration_id`) | Delete the field from HCL |
| `ngrok_service_user` (unchanged name, was `ngrok_service_user` in v1) | No change needed — resource name is preserved |
| Schema type changes (SDKv2 `TypeSet` → Framework nested objects) | Handled automatically by state upgraders where possible |

### 9.3 Upgrade Guide

Publish a step-by-step upgrade guide in the Terraform Registry docs (`docs/guides/version-1-upgrade.md`) covering:

1. Update `required_providers` version constraint to `~> 1.0`
2. Run `terraform init -upgrade`
3. Remove deprecated resource blocks and `terraform state rm` them
4. Remove deprecated fields from carried-over resources
5. Run `terraform plan` to verify — no infrastructure changes should occur for carried-over resources
6. Replace edge/backend workflows with `ngrok_cloud_endpoint` + traffic policy

---

## 10. Open Questions

1. **OpenAPI spec completeness** — Need to audit whether `ngrok.yaml` covers all API resources or if some are undocumented.

2. **Multi-account support** — Should the provider support aliased provider configurations for managing resources across multiple ngrok accounts?

3. **Rate limiting** — The `ngrok-api-go` client may need retry/backoff configuration exposed as provider settings.

4. **Terraform Cloud / HCP integration** — Any special considerations for running in Terraform Cloud (e.g., environment variable handling)?
