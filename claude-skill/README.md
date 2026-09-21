# moxy skill for Claude Code

A [Claude Code](https://claude.com/claude-code) skill that lets you set up mocks and proxies
by asking, instead of hand-editing `mockdef.json`:

> mock `GET /api/users/123` as a 500
>
> proxy `/api/orders` to localhost:8080
>
> what am I currently mocking?

Claude reads `SKILL.md` for moxy's matching rules and drives `scripts/moxyctl.py`, which
talks to the admin HTTP API on port 9097. Because it goes over HTTP rather than editing
files, it works from **any** project directory — you don't have to be in the moxy repo, and
Claude never needs to know where moxy is installed.

## Install

The skill has to live in `~/.claude/skills/` for Claude Code to find it. Symlink it, so it
stays in sync when you pull. From the repo root:

```sh
mkdir -p ~/.claude/skills
rm -rf ~/.claude/skills/moxy          # a leftover directory would swallow the symlink
ln -s "$(pwd)/claude-skill" ~/.claude/skills/moxy
```

Or copy it, if you'd rather it not move when the repo does:

```sh
mkdir -p ~/.claude/skills/moxy
cp -r claude-skill/. ~/.claude/skills/moxy/
```

Either way the skill ends up at `~/.claude/skills/moxy/SKILL.md`.

**Restart Claude Code afterwards** — skills are scanned at startup, so a freshly installed
one is not picked up mid-session.

To install it for a single project instead of your whole account, put it in that project's
`.claude/skills/moxy/` rather than `~/.claude/skills/moxy/`.

## Check it works

Start moxy, then ask Claude something like *"what mocks are active right now?"*. You can
also run the script yourself:

```sh
python3 ~/.claude/skills/moxy/scripts/moxyctl.py list
```

Requires python3 (standard library only) and a running moxy. If moxy listens on a port other
than 9097, pass `--port 9098` before the command, or export `MOXY_PORT=9098`.

## What the script can do

```
list                    mocks and proxies, in match order, with their indices
raw mock|proxy          the full JSON for one of the two definition files
add-mock                --urlpart --method --status --payload --freeze --comment
add-proxy               --urlpart --target --verbose --comment
set-active mock|proxy on|off   --index N | --urlpart TEXT
remove mock|proxy       --index N | --urlpart TEXT
replace mock|proxy      --file whole-file.json
```

New entries are added at the **top** by default, since moxy matches top to bottom and stops
at the first hit; `--at-bottom` puts one last instead.

Writes go through moxy's admin API, which validates the JSON before saving. A malformed
definition is rejected with a `400` and the file is left untouched — worth knowing, because a
broken `mockdef.json` otherwise makes moxy answer `500` to *every* request until it's fixed.
