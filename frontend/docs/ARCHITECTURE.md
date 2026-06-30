# Frontend Architecture & Screen Map

How the Cloud Log Access SPA is organized: every route and screen, the UI states each screen can
be in, and the layered structure (API → hooks → stores → components) that powers them.

- **Setup, screenshots, design decisions:** [`../README.md`](../README.md)
- **As-built API contract:** [`../../docs/FRONTEND-INTEGRATION.md`](../../docs/FRONTEND-INTEGRATION.md)
- **Original architecture plan:** [`../../docs/FRONTEND-PLAN.md`](../../docs/FRONTEND-PLAN.md)

---

## 1. At a glance

```
                         ┌──────────────────────────────────────────────┐
  Browser  ── /api/v1 ─► │  nginx (prod) / Vite proxy (dev)  ── same-origin ──► Go BFF
                         └──────────────────────────────────────────────┘
        │
        ▼  React 18 SPA
  ┌──────────────────────────────────────────────────────────────────────────┐
  │  pages/      LoginPage · DashboardPage · LogsPage     (route entry points) │
  │     │ consume                                                              │
  │  components/ ui · layout · logs · dashboard           (presentational)     │
  │     │ data via                                                            │
  │  hooks/      useProviders · useLogs · useDownload · useTempLink · useMe …  │
  │     │ call               │ read/write                                     │
  │  api/        client + endpoint modules        store/  auth · ui · toast    │
  │     │ over fetch                                (Zustand, some persisted)   │
  │  lib/        cn · format · theme · constants   styles/ token layer (CSS)   │
  └──────────────────────────────────────────────────────────────────────────┘
```

Three layers do the work:
- **Server state** → TanStack Query hooks (`hooks/`) calling the typed API client (`api/`).
- **Global client state** → three Zustand stores (`store/`): session, UI prefs, transient toast.
- **Presentation** → `pages/` (route-level orchestration) built from `components/` primitives.

---

## 2. Routes

Defined in [`router/index.tsx`](../src/router/index.tsx); the auth gate is
[`router/ProtectedRoute.tsx`](../src/router/ProtectedRoute.tsx).

| Path | Guard | Renders | Purpose |
|---|---|---|---|
| `/login` | public | `LoginPage` | Authenticate (email + password) |
| `/` | auth | → redirect to `/dashboard` | — |
| `/dashboard` | auth | `AppShell` → `DashboardPage` | Overview: live multi-cloud health + file stats |
| `/logs` | auth | `AppShell` → `LogsPage` | Browse / search / download logs; temp-links |
| `*` | — | → redirect to `/dashboard` | Catch-all |

`ProtectedRoute` redirects unauthenticated users to `/login` (preserving `from`), and supports an
optional `requiredRole` (role-guarded routes redirect mismatches to `/dashboard`). RBAC today is
enforced at the **action/provider** level inside the Logs screen rather than via separate routes.

---

## 3. Screen map

### 3.1 Login — `pages/LoginPage.tsx`

Full-viewport centered card. Not wrapped in `AppShell`.

| Region | Component / source |
|---|---|
| Brand + subtitle | inline (lucide `CloudRain` in `--primary`) |
| Email / Password fields | `ui/Input` + React Hook Form + Zod (`.email()`, password required) |
| Show/hide password | inline toggle button (`aria-pressed`) |
| Submit | `ui/Button` (`loading` while the mutation is pending) |
| Demo-creds hint | `lib/constants` → `DEMO_CREDENTIALS` |
| Theme control | `ui/ThemeToggle variant="pill"` (Light / Dark / System) |

**States**

| State | Trigger | UI |
|---|---|---|
| idle | first paint | empty form |
| field error | Zod fails on submit | `role="alert"` message under the field, `aria-invalid` |
| submitting | mutation pending | button spinner + `aria-busy`, inputs stay editable |
| credentials error | `401 INVALID_CREDENTIALS` | inline "Invalid email or password." (no redirect) |
| network/other error | any other error | inline `err.message` fallback |
| session-expired banner | `?reason=expired` in URL | `role="status"` banner above the card |
| success | `200` | `authStore.login()` → navigate to `from` or `/dashboard` |
| already authed | mount with a session | `<Navigate to="/dashboard" replace/>` |

Screenshots: `01-login.png`, `11-login-light.png`.

### 3.2 App shell (Dashboard + Logs share it) — `components/layout/`

```
AppShell
├── SkipLink                         "Skip to content" → #main
├── Sidebar       228px sticky       brand · NAVIGATION · Dashboard/Logs nav · Logout
├── Topbar        56px sticky        Breadcrumb (left) · ThemeToggle(icon) · UserChip (right)
│   ├── Breadcrumb                   "Dashboard"  |  "Logs / {provider} / {bucket}"
│   └── UserChip → RoleBadge         avatar · email local-part · Admin/Viewer badge
└── <Outlet/> inside <main id="main">
Toast (portal)                        bottom-right download feedback
```

- On mount, `AppShell` calls `useMe()` to **re-validate the persisted session** and refresh
  `allowed_providers`; a revoked token `401`s and is caught by the global redirect.
- `Sidebar` logout: `POST /auth/logout` (best-effort) → clear `authStore` → navigate `/login`.
- `Breadcrumb` omits the bucket segment when the current provider is **not** permitted (no-permission
  showing) — it never leaks a disallowed bucket name.

### 3.3 Dashboard — `pages/DashboardPage.tsx`

Composed from **real endpoints** (`/providers` + the primary bucket's `/logs`); there is no
`/summary` endpoint on the BFF.

| Region | Component | Data |
|---|---|---|
| 4 stat tiles | `dashboard/StatCard` | provider name+region · primary bucket · file count · total size |
| Recent log files | `dashboard/RecentFiles` | newest files from `/logs` (level badge, size, modified) |
| Providers | `dashboard/ProvidersPanel` | every provider from `/providers` with a live health badge |

**States:** `loading` (skeleton tiles + rows) · `error` (providers fetch failed → inline card) ·
`empty` (no providers for the account) · `content`. Screenshots: `02-dashboard.png`,
`12-dashboard-light.png`, `14-prod-container.png`.

### 3.4 Logs — `pages/LogsPage.tsx` (the primary screen)

The page is a **state machine** that renders exactly one mutually-exclusive surface below a fixed
header + toolbar. Two view variants (Table / Cards) share the same data and actions.

```
Header:  "Logs" + subtitle (bucket)            ViewToggle (Table|Cards) + Reload
Toolbar: Provider <select>                     Search input (live client-side filter)
─────────────────────────────────────────────────────────────────────────────────
            exactly one of ↓
   NoPermissionState · SkeletonTable · ErrorState · EmptyState · (LogsTable | LogsCards)
```

**Toolbar & header components** (`components/logs/`): `LogsToolbar`, `ProviderSelect` (shows all 3,
marks disallowed ones "(no access)"), `SearchInput`, `ViewToggle`.

**Data components:** `LogsTable` → `LogsTableRow`; `LogsCards` → `LogCard`; `LogRowActions`
(admin-only temp-link button); `TempLinkModal` (admin-only, presigned URL).

#### State-machine inputs

```ts
isProviderAllowed = !user.allowed_providers || allowed_providers.includes(provider)
providerUnhealthy = providers.find(p => p.id===provider)?.healthy === false   // fast fail signal
bucket            = providerObj?.buckets[0] ?? ''
q                 = useLogs(provider, bucket, isProviderAllowed && !providerUnhealthy)
is403             = q.error is ApiResponseError && code === 'PROVIDER_NOT_ALLOWED'
```

| Surface | Condition | Component | Notes |
|---|---|---|---|
| **No permission** | `!isProviderAllowed || is403` | `NoPermissionState` | amber lock; **no fetch** (preemptive) + defensive 403; "Switch to {first allowed}" |
| **Skeleton** | `q.isLoading || providersLoading` | `SkeletonTable` | shimmer rows |
| **Error** | `providerUnhealthy || (q.isError && !is403) || noBucket` | `ErrorState` | `503`/unreachable; **Retry** re-probes `/providers` |
| **Empty** | `q.isSuccess && filtered.length === 0` | `EmptyState` | "no files" or "no match \"{q}\"" + Clear search |
| **Data** | `q.isSuccess && filtered.length > 0` | `LogsTable` / `LogsCards` | by `uiStore.logsView` |

#### Actions

- **Download** (both roles): `useDownload()` → `client.downloadToBlob` (Bearer fetch → Blob → hidden
  anchor). Per-row spinner via `download.isPending && download.variables?.key === file.key`. Drives the
  `Toast` (progress → done / error).
- **Temporary link** (admin only): `onLink` opens `TempLinkModal` (local `useState`). The link button
  is **DOM-absent** for viewers (`useIsAdmin()`), so the action cannot be triggered. The modal calls
  `useTempLink()` → `POST /presign`; the presigned URL is **copied / navigated, never `fetch`ed**.

Screenshots: `03-logs-table.png`, `04-logs-cards.png`, `05/06-temp-link-*.png`, `07-error-state.png`,
`08-empty-state.png`, `09-viewer-logs-table.png`, `10-no-permission.png`, `13-logs-azure.png`.

---

## 4. Screen × state matrix (QA checklist)

| Screen | loading | populated | empty | error | no-permission(403) | auth-expired(401) |
|---|---|---|---|---|---|---|
| Login | submit spinner | — | — | inline `role=alert` | — | `?reason=expired` banner |
| App shell | — | nav active + real user/role | — | global 401 → redirect | — | global 401 → redirect |
| Dashboard | skeletons | stats + recent + providers | "no providers" card | inline card | — | global 401 → redirect |
| Logs table/cards | shimmer | rows / cards | "No log files" / "No match" | "Unable to connect" + Retry | amber lock + Switch | global 401 → redirect |
| Download | toast "Downloading…" | — | — | toast `role=alert` | (allowed any role) | global 401 → redirect |
| Temp-link (admin) | modal generate spinner | link + Copy + expiry helper | — | inline modal error | viewer never sees trigger | global 401 → redirect |

**Per-cell gate:** keyboard-reachable · focus-visible ring · SR label/announce · reduced-motion ·
AA contrast · 200% reflow.

---

## 5. Project organization

```
frontend/
├── index.html              # no-flash theme <script> (CSP sha256), font preconnect
├── nginx.conf              # prod: SPA fallback + /api proxy + security headers (CSP/XFO/…)
├── Dockerfile              # node build → nginx-unprivileged serve
├── scripts/csp-hash.mjs    # recompute the inline-script CSP hash at build
└── src/
    ├── main.tsx            # StrictMode + QueryClient + Router; theme sync; optional MSW boot
    │
    ├── api/                # ── transport layer (typed fetch over /api/v1) ──
    │   ├── client.ts       #   request<T> (envelope unwrap, 401 routing), downloadToBlob, ApiResponseError
    │   ├── auth.ts         #   login · logout · me
    │   ├── providers.ts    #   listProviders
    │   ├── logs.ts         #   listLogs · downloadPath
    │   ├── presign.ts      #   createPresignedLink            (admin)
    │   └── types.ts        #   DTO mirror of the Go structs (snake_case, authoritative)
    │
    ├── store/              # ── global client state (Zustand) ──
    │   ├── authStore.ts    #   token/user/isAuthenticated; persisted 'cla_session'; cross-tab sync
    │   ├── uiStore.ts      #   theme · logsView · provider · search (view+provider persisted 'cla_ui')
    │   └── toastStore.ts   #   transient download status (not persisted)
    │
    ├── hooks/              # ── server state + actions (TanStack Query) ──
    │   ├── useProviders · useLogs · useMe                                  (queries)
    │   ├── useDownload · useTempLink                                       (mutations)
    │   └── useIsAdmin · useMediaQuery                                      (selectors / helpers)
    │
    ├── router/             # createBrowserRouter + ProtectedRoute (auth + optional role guard)
    │
    ├── pages/              # route entry points: LoginPage · DashboardPage · LogsPage
    │
    ├── components/
    │   ├── ui/             # ── design-system primitives (no business logic) ──
    │   │                   #   Button IconButton Input Select Badge Card SegmentedControl
    │   │                   #   Spinner Skeleton SkeletonTable Modal Toast ThemeToggle SkipLink
    │   │                   #   EmptyState (+ ErrorState, NoPermissionState specializations)
    │   ├── layout/         # AppShell Sidebar Topbar Breadcrumb UserChip RoleBadge
    │   ├── logs/           # LogsToolbar ProviderSelect SearchInput ViewToggle
    │   │                   #   LogsTable/Row LogsCards/Card LogRowActions TempLinkModal
    │   └── dashboard/      # StatCard RecentFiles ProvidersPanel
    │
    ├── lib/                # cn (class merge) · format (bytes/time/level) · theme · constants
    ├── styles/             # tokens.css · keyframes.css · base.css · index.css (orchestrator)
    └── mocks/              # MSW handlers + worker (dev/demo only; VITE_USE_MOCKS=true)
```

**Layering rule:** dependencies point downward only — `pages → components → hooks → (api | store) →
lib`. Components never call `fetch` directly; hooks never render; the API client is the only place
that knows about the wire format.

---

## 6. State management

| Concern | Where | Persisted? |
|---|---|---|
| Auth session (`token`, `user`, `isAuthenticated`) | `authStore` (Zustand) | ✅ `cla_session` (+ cross-tab `storage` sync) |
| Theme preference | `lib/theme` (`cla_theme`) mirrored into `uiStore.theme` | ✅ `cla_theme` (read by the no-flash script) |
| Logs view + selected provider | `uiStore` | ✅ `cla_ui` (partialized) |
| Search query | `uiStore.search` | ❌ ephemeral |
| Download toast | `toastStore` | ❌ ephemeral |
| Server data (providers, logs, me) | TanStack Query cache | ❌ (in-memory, `staleTime`-bounded) |
| Temp-link modal lifecycle | `LogsPage` local `useState` | ❌ (single-screen concern) |

**Query keys & policy:** `['providers']` (60s stale) · `['logs', provider, bucket]` (30s stale,
`retry:false` — manual Retry) · `['me']` (5min stale). The logs query is **disabled** unless the
provider is allowed, healthy, and a bucket is resolved.

---

## 7. Cross-cutting concerns

**Error routing** (`api/client.ts` + `LogsPage`) — branch on `error.code`, not just HTTP status:

| Code(s) | Where handled | UI |
|---|---|---|
| `INVALID_CREDENTIALS` | login mutation `onError` | inline form error (no redirect) |
| `TOKEN_EXPIRED` / `TOKEN_INVALID` / `TOKEN_MISSING` | `client.request` (global) | hard-redirect `/login?reason=expired` |
| `PROVIDER_NOT_ALLOWED` (403) | `LogsPage` | `NoPermissionState` |
| `UNAUTHORIZED` (403, admin route) | viewer never reaches it | n/a (action is DOM-absent) |
| `STORAGE_UNAVAILABLE` (503) / unhealthy | `LogsPage` | `ErrorState` + Retry |

**RBAC:** `user.allowed_providers` (from the login body, also refreshed via `/me`) drives the provider
gate; `useIsAdmin()` gates the temp-link affordance. The UI mirrors the backend's authority for UX —
the backend remains the enforcement point.

**Theme system:** one CSS-variable token layer (`styles/tokens.css`) swapped by `data-theme` on
`<html>`. `light` / `dark` / `system`; an inline `<head>` script resolves it before first paint
(no flash) and is CSP-whitelisted by sha256. `--primary: #059669` is the single re-skin token; a
`*-strong` family provides AA-passing accent text.

**Accessibility:** real semantic `<table>` with `<th scope="row">`; labelled icon-only buttons;
focus-trapped `role="dialog"` modal (Esc / click-outside / focus-return); `role="status"`/`alert`
toasts; arrow-key `radiogroup` theme controls; `:focus-visible` rings; `prefers-reduced-motion`
handling; skip link → focusable `#main`.

---

## 8. Request lifecycle (worked example)

A download, end to end:

```
1. User clicks Download in LogsTableRow
2. onDownload(file) → useDownload().mutate({ provider, bucket, key })
3. onMutate         → toastStore.show(filename)            (toast: "Downloading…")
4. mutationFn       → client.downloadToBlob(downloadPath(...), filename)
5. client           → fetch(/api/v1/providers/{p}/download?…) with Authorization: Bearer
                       └─ nginx (prod) / Vite proxy (dev) → BFF streams the object
6. response.blob()  → object URL → hidden <a download> click → revokeObjectURL
7. onSuccess        → toastStore.complete(filename)        (toast: green check, auto-dismiss)
   onError(401)     → client redirects to /login?reason=expired
   onError(other)   → toastStore.error(filename)           (toast: role=alert)
```

The same shape repeats everywhere: **component → hook → api client → (proxy) → BFF**, with results
flowing back through the Query cache / stores into the mutually-exclusive UI surfaces above.
