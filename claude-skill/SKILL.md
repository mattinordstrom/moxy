---
name: moxy
description: Inspect and change HTTP mocks and proxies in moxy, the local mocking/proxying tool running on port 9097. Use whenever the user wants to mock or stub an API endpoint, fake a JSON response, simulate an error status code or a slow/timing-out response, route localhost traffic to a different backend, or list/enable/disable what is currently mocked or proxied.
---

# moxy

moxy is a local HTTP mock + proxy, listening on **port 9097** by default. The app under
development points at `http://localhost:9097`; moxy either answers with a canned mock or
forwards the request on to the real backend.

Drive it through the admin HTTP API, not by editing files on disk — the API works from any
working directory and keeps the on-disk formatting identical to what the admin UI writes.

## The helper script

```bash
python3 ~/.claude/skills/moxy/scripts/moxyctl.py <command>
```

It talks to moxy over HTTP, so it works from any project directory. If moxy runs on a
different port, pass `--port` before the command (or set `MOXY_PORT`).

Always run `list` first — order matters, and an existing entry may already shadow what you
are about to add.

```bash
moxyctl.py list                       # all mocks and proxies with their indices
moxyctl.py raw mock|proxy             # full JSON for one of the two files

moxyctl.py add-mock --urlpart /api/users/123 --method GET --status 200 \
    --payload '{"id":123,"name":"Anna"}'
moxyctl.py add-mock --urlpart /api/report --payload-file ./big.json   # inline a local file
moxyctl.py add-mock --urlpart /api/report --payload-from-file /abs/path/on/disk.json
moxyctl.py add-mock --urlpart /api/slow --freeze 30000                # simulate a timeout
moxyctl.py add-mock --urlpart /api/boom --status 500 --payload '{"error":"nope"}'

moxyctl.py add-proxy --urlpart /api/orders --target http://localhost:8080 --verbose
moxyctl.py add-proxy --urlpart /api/common --target http://route-utv.fnox.se --preserve-host

moxyctl.py set-active mock off --urlpart /api/users   # disable without deleting
moxyctl.py set-active proxy on --index 0
moxyctl.py remove mock --urlpart /api/users
moxyctl.py replace mock --file ./new-mockdef.json     # overwrite the whole list
```

New entries go to the **top** of the list by default, so they win over anything already
there; pass `--at-bottom` for a low-priority catch-all. `--comment` is optional and shows up
in the admin UI. Prefer `set-active ... off` over `remove` — the user often wants a mock back.

## Matching rules

These are the rules the request actually follows, and most surprises come from them:

- **Mocks are checked before proxies**, both top to bottom, **first match wins**. A broad
  entry high in the list shadows everything more specific below it.
- A mock matches when it is `active`, the **method matches exactly**, and the request URL
  **contains** the `urlpart` (leading/trailing `/` are stripped before comparing). So
  `urlpart: "api/users"` also matches `/api/users/123/settings`. Be as precise as the user's
  intent requires.
- The compared URL includes the **query string**, so `urlpart: "/api/search?type=full"` works.
- If `urlpart` contains `.*` it is also tried as a **regex** against the full URL:
  `"api/users/.*/settings"`.
- Proxies match the same way but **ignore the HTTP method**; the request is forwarded to
  `target` + the original path.
- No proxy match → the request goes to the default route in `config.yml`.
- `payload` beats `payloadFromFile`; the file is only read when `payload` is null or empty.
  `payloadFromFile` must be an **absolute path** and is read by moxy, not by you.
- Mock responses are always sent as `application/json`.

## Two things that will bite you

1. **Changes need no restart, but they take ~1.5s.** The definition files are cached for
   1500ms, and the write itself refreshes that cache, so a request fired immediately after a
   change still gets the *old* behaviour. Wait ~2s before testing, and don't conclude a mock
   is broken from one fast curl. (Verified: a fresh mock started answering at 1.5s.)
2. **An open admin UI tab refreshes itself.** moxy watches the definition files and pushes a
   `defschanged` event over its websocket, so the UI re-renders within ~1s of any change —
   yours, a hand edit, anything. The one exception is a field holding unsaved keystrokes: the
   reload is deferred until that field loses focus, so typing is never destroyed. Both cases
   are logged in the UI (`reloaded` / `will reload when you finish editing`), so if the user
   reports a stale UI, ask what the log says before doubting the watcher.

   If they see *nothing* in the log and the list is stale, suspect a **stale cached
   `main.js`** rather than the server. moxy sends `Cache-Control: no-cache` for
   `/ui/static/`, but a browser that cached those scripts before that header existed can keep
   serving the old build — a hard reload (Ctrl+Shift+R) clears it. Editing UI js needs no
   moxy restart (it is served from disk per request); only Go changes do.

## Verifying

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:9097/api/users/123
```

moxy logs every hit: mocks in purple, proxied requests in green, default-route in grey —
both in its terminal and in the admin UI log.

## What moxy cannot do

Don't invent fields for these; say they aren't supported and offer an alternative:
custom response headers, matching on request body or headers, per-request sequenced
responses, and dynamic/templated payloads. A mock is one fixed body per `urlpart` + method.

## Entry shapes

`mockdef.json`:

```json
{ "active": true, "comment": "optional", "freezetimems": 0, "method": "GET",
  "payload": { "any": "json" }, "payloadFromFile": "", "statuscode": 200,
  "urlpart": "/api/whatever" }
```

`proxydef.json`:

```json
{ "active": true, "comment": "optional", "preserveHost": false,
  "target": "http://localhost:8080", "urlpart": "/api/whatever", "verbose": false }
```

`verbose` on a proxy logs headers and body and enables "copy as curl" in the admin UI.

`preserveHost` forwards the incoming `Host` header instead of the target's. Use it when the
target routes on Host (an ingress like route-utv) and the client sets Host explicitly, e.g. a
Java backend whose `httpClient` points at moxy. A loopback Host (`localhost:9097`) is never
forwarded, so it is harmless for browsers. The default route has the same switch as
`defaults.preserveHost` in moxy's `config.yml`, which needs a moxy restart; ask the user to
change it rather than editing that file. Symptom of a missing `preserveHost`: the target
answers a plain-text `404 page not found` for a path that works when called directly.
