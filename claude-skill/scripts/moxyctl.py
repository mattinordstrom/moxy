#!/usr/bin/env python3
import argparse
import json
import os
import sys
import urllib.error
import urllib.request

DEFAULT_PORT = os.environ.get("MOXY_PORT", "9097")
BASE = f"http://localhost:{DEFAULT_PORT}/moxyadminui"


def fail(msg):
    print(f"error: {msg}", file=sys.stderr)
    sys.exit(1)


def get_defs(kind):
    url = f"{BASE}/{kind}def"
    try:
        with urllib.request.urlopen(url, timeout=10) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.URLError as exc:
        fail(f"cannot reach moxy at {url} ({exc}). Is moxy running there?")


def put_defs(kind, data):
    url = f"{BASE}/{kind}def"
    req = urllib.request.Request(
        url,
        data=json.dumps(data, ensure_ascii=False).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            resp.read()
    except urllib.error.HTTPError as exc:
        fail(f"moxy rejected the write: {exc.code} {exc.read().decode().strip()}")
    except urllib.error.URLError as exc:
        fail(f"cannot reach moxy at {url} ({exc}). Is moxy running there?")


def describe_mock(idx, m):
    flag = "on " if m.get("active") else "off"
    freeze = f" {m['freezetimems']}ms" if m.get("freezetimems") else ""
    src = "file:" + m["payloadFromFile"] if not m.get("payload") and m.get("payloadFromFile") else "inline"
    comment = f"  # {m['comment']}" if m.get("comment") else ""
    return f"[{idx}] {flag} {m.get('method','?'):6} {m.get('statuscode','?')}{freeze} {m.get('urlpart','')} ({src}){comment}"


def describe_proxy(idx, p):
    flag = "on " if p.get("active") else "off"
    verbose = " verbose" if p.get("verbose") else ""
    comment = f"  # {p['comment']}" if p.get("comment") else ""
    return f"[{idx}] {flag} {p.get('urlpart','')} -> {p.get('target','')}{verbose}{comment}"


def cmd_list(args):
    mocks = get_defs("mock")
    proxies = get_defs("proxy")
    print("MOCKS (matched first, top to bottom):")
    for i, m in enumerate(mocks):
        print("  " + describe_mock(i, m))
    if not mocks:
        print("  (none)")
    print("\nPROXIES (matched only if no mock matched):")
    for i, p in enumerate(proxies):
        print("  " + describe_proxy(i, p))
    if not proxies:
        print("  (none)")


def cmd_raw(args):
    print(json.dumps(get_defs(args.kind), indent=2, ensure_ascii=False))


def resolve_payload(args):
    if args.payload is not None:
        try:
            return json.loads(args.payload)
        except json.JSONDecodeError as exc:
            fail(f"--payload is not valid JSON: {exc}")
    if args.payload_file is not None:
        try:
            with open(args.payload_file) as fh:
                return json.load(fh)
        except (OSError, json.JSONDecodeError) as exc:
            fail(f"cannot read --payload-file: {exc}")
    return None


def cmd_add_mock(args):
    payload = resolve_payload(args)
    if payload is None and not args.payload_from_file:
        fail("give one of --payload, --payload-file or --payload-from-file")
    if args.payload_from_file and not args.payload_from_file.startswith("/"):
        fail("--payload-from-file must be an absolute path (moxy reads it server-side)")

    entry = {
        "active": True,
        "freezetimems": args.freeze,
        "method": args.method.upper(),
        "payload": payload,
        "payloadFromFile": args.payload_from_file or "",
        "statuscode": args.status,
        "urlpart": args.urlpart,
    }
    if args.comment:
        entry["comment"] = args.comment

    mocks = get_defs("mock")
    if args.at_bottom:
        mocks.append(entry)
        idx = len(mocks) - 1
    else:
        mocks.insert(0, entry)
        idx = 0
    put_defs("mock", mocks)
    print("added " + describe_mock(idx, entry))


def cmd_add_proxy(args):
    entry = {
        "active": True,
        "target": args.target.rstrip("/"),
        "urlpart": args.urlpart,
        "verbose": args.verbose,
    }
    if args.comment:
        entry["comment"] = args.comment

    proxies = get_defs("proxy")
    if args.at_bottom:
        proxies.append(entry)
        idx = len(proxies) - 1
    else:
        proxies.insert(0, entry)
        idx = 0
    put_defs("proxy", proxies)
    print("added " + describe_proxy(idx, entry))


def pick(items, kind, index, urlpart):
    if index is not None:
        if index < 0 or index >= len(items):
            fail(f"no {kind} at index {index} (have {len(items)})")
        return [index]
    matches = [i for i, it in enumerate(items) if urlpart in it.get("urlpart", "")]
    if not matches:
        fail(f"no {kind} whose urlpart contains {urlpart!r}")
    return matches


def cmd_set_active(args):
    items = get_defs(args.kind)
    describe = describe_mock if args.kind == "mock" else describe_proxy
    for i in pick(items, args.kind, args.index, args.urlpart):
        items[i]["active"] = args.state == "on"
    put_defs(args.kind, items)
    for i in pick(items, args.kind, args.index, args.urlpart):
        print(describe(i, items[i]))


def cmd_remove(args):
    items = get_defs(args.kind)
    describe = describe_mock if args.kind == "mock" else describe_proxy
    targets = pick(items, args.kind, args.index, args.urlpart)
    removed = [describe(i, items[i]) for i in targets]
    items = [it for i, it in enumerate(items) if i not in set(targets)]
    put_defs(args.kind, items)
    for line in removed:
        print("removed " + line)


def cmd_replace(args):
    try:
        with open(args.file) as fh:
            data = json.load(fh)
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"cannot read {args.file}: {exc}")
    if not isinstance(data, list):
        fail("the file must contain a JSON array")
    put_defs(args.kind, data)
    print(f"wrote {len(data)} {args.kind} entries")


def main():
    parser = argparse.ArgumentParser(prog="moxyctl")
    parser.add_argument("--port", default=DEFAULT_PORT, help="moxy port (default 9097, or $MOXY_PORT)")
    sub = parser.add_subparsers(dest="cmd", required=True)

    sub.add_parser("list").set_defaults(func=cmd_list)

    p = sub.add_parser("raw")
    p.add_argument("kind", choices=["mock", "proxy"])
    p.set_defaults(func=cmd_raw)

    p = sub.add_parser("add-mock")
    p.add_argument("--urlpart", required=True)
    p.add_argument("--method", default="GET")
    p.add_argument("--status", type=int, default=200)
    p.add_argument("--payload")
    p.add_argument("--payload-file")
    p.add_argument("--payload-from-file")
    p.add_argument("--freeze", type=int, default=0)
    p.add_argument("--comment")
    p.add_argument("--at-bottom", action="store_true")
    p.set_defaults(func=cmd_add_mock)

    p = sub.add_parser("add-proxy")
    p.add_argument("--urlpart", required=True)
    p.add_argument("--target", required=True)
    p.add_argument("--verbose", action="store_true")
    p.add_argument("--comment")
    p.add_argument("--at-bottom", action="store_true")
    p.set_defaults(func=cmd_add_proxy)

    p = sub.add_parser("set-active")
    p.add_argument("kind", choices=["mock", "proxy"])
    p.add_argument("state", choices=["on", "off"])
    p.add_argument("--index", type=int)
    p.add_argument("--urlpart")
    p.set_defaults(func=cmd_set_active)

    p = sub.add_parser("remove")
    p.add_argument("kind", choices=["mock", "proxy"])
    p.add_argument("--index", type=int)
    p.add_argument("--urlpart")
    p.set_defaults(func=cmd_remove)

    p = sub.add_parser("replace")
    p.add_argument("kind", choices=["mock", "proxy"])
    p.add_argument("--file", required=True)
    p.set_defaults(func=cmd_replace)

    args = parser.parse_args()

    global BASE
    BASE = f"http://localhost:{args.port}/moxyadminui"

    if args.cmd in ("set-active", "remove") and args.index is None and not args.urlpart:
        fail("give either --index or --urlpart")
    args.func(args)


if __name__ == "__main__":
    main()
