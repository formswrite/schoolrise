# Building the SchoolRise Encore-Go fork

The `schoolrise-app:patched` Docker image is the most fragile part of the deployment story. This page walks from `git clone encore-go-fork` to a built image that can run BAML-backed services on Linux/aarch64 Docker.

> **Why this exists.** Encore's stock CLI builds binaries with `StaticLink: true`. BAML uses cgo with a memory layout that overflows the aarch64 glibc static-TLS budget under Docker on Apple Silicon. The result is a SIGSEGV the moment any BAML function is called. We patch two lines in Encore's build pipeline to disable static linking; everything else (BAML, our services, the rest of the toolchain) is unmodified.
>
> Once GitHub Actions builds + pushes the image to `ghcr.io/formswrite/schoolrise-app` (production-roadmap item #9), most contributors will never need to do this. Until then, you build it locally.

---

## What we patch in Encore

Two files in our fork of `encoredotdev/encore`:

| File | Change |
|---|---|
| `cli/daemon/export/export.go` | `StaticLink: true` → `false` (export builds match runtime builds) |
| `v2/compiler/build/build.go` | Wrap the static-link branch in `if build.CgoEnabled { env = append(env, "CGO_ENABLED=1") } else if !build.CgoEnabled { ... }` so cgo-marked services emit `CGO_ENABLED=1` to the underlying `go build` |

That's it. Two diffs, ~6 lines total. Everything else in Encore is upstream.

---

## Prerequisites

- **macOS Apple Silicon or Linux** with Docker installed
- **Go 1.23+** (Encore's CLI builds against this)
- **Git** with submodule support
- **~3 GB free disk** (the Encore CLI build artefacts are ~2.5 GB)
- The `encore` CLI installed once (`brew install encoredev/tap/encore`) so we have somewhere to install the patched binary

---

## Step 1 — Clone our Encore fork

```bash
# In the directory above schoolrise/
git clone https://github.com/formswrite/encore-go.git
cd encore-go
git checkout schoolrise-static-link-fix
```

The branch `schoolrise-static-link-fix` is rebased on top of upstream `encoredotdev/encore` periodically. Check `git log upstream/main..HEAD` to see what we've added.

---

## Step 2 — Build the patched encore CLI

```bash
# Inside encore-go/
make build
# Produces ./encore (~120 MB binary)

# Install it over the brew-installed one
mv "$(brew --prefix encoredev/tap/encore)/bin/encore" "$(brew --prefix encoredev/tap/encore)/bin/encore.brew-original"
ln -s "$PWD/encore" "$(brew --prefix encoredev/tap/encore)/bin/encore"

# Sanity check
encore version
# Expect: encore version v1.x.y (schoolrise-static-link-fix)
```

To revert later: remove the symlink and rename `encore.brew-original` back.

---

## Step 3 — Build the SchoolRise app image with the patched CLI

From the schoolrise repo root:

```bash
cd /path/to/schoolrise

# This is the long step — first run is ~8 minutes; subsequent runs use Docker layer cache (~90 s)
encore build docker schoolrise-app:rebuilt-arm64 \
    --config infra-config/selfhost.json

# `encore build` produces a "headless" runtime image. We add a thin wrapper layer
# that mounts the BAML cache directory at /baml-cache so the LLM contracts work.
docker build -f deploy/Dockerfile.app-wrapper -t schoolrise-app:patched .
```

Verify:

```bash
docker images | grep schoolrise-app:patched
# Expect: ~150 MB image
```

---

## Step 4 — Smoke test

```bash
# Start the local stack (docker-compose.local.yml uses schoolrise-app:patched)
make compose-up-local

# Wait for the app to be ready
sleep 15

# Hit a BAML-backed endpoint to confirm cgo works
curl -s http://localhost:8080/v1/ai/health
# Expect: {"status":"ok"}
```

If you see a SIGSEGV in `docker logs deploy-app-1` instead, the patch didn't apply. Re-run step 2 and verify `encore version` shows your fork.

---

## Step 5 — Updating the fork

We rebase on upstream `encoredotdev/encore` periodically:

```bash
cd encore-go
git fetch upstream
git rebase upstream/main schoolrise-static-link-fix

# Resolve any conflicts in the two patched files
# Re-run `make build` to confirm

git push origin schoolrise-static-link-fix --force-with-lease
```

Document the rebase in [CHANGELOG.md](../CHANGELOG.md) under "Encore upstream sync".

---

## Why this isn't fixed upstream

We've discussed this with the Encore maintainers. The static-link default is a deliberate choice for their hosted product (smaller binaries, no shared-lib runtime deps). For self-hosted users with cgo-using libraries (BAML, sqlite, etc.), the default doesn't fit. Upstream may eventually expose a `--no-static-link` flag — until then, the fork.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `encore build` complains "BAML library not found" | The wrapper-layer Dockerfile expects `~/.baml/cache`. If empty, the wrapper has nothing to copy. | Run `baml-cli generate` once in `apps/ai/` before the docker build, or pull a prebuilt cache from CI. |
| Container crashes with `cannot allocate memory in static TLS block` | The patch didn't apply (your `encore` binary is the brew one, not the fork). | `which encore` should print the symlinked fork path. If it shows the brew path, redo step 2. |
| `make compose-up-local` errors with "image not found" | `schoolrise-app:patched` was pruned by Docker. | Re-run step 3. |
| Build takes >15 min | First-run image-layer cache miss. | Subsequent runs are fast; if it's persistently slow, check `docker system df` for cache pressure. |

---

## What this story looks like once CI is wired

The plan: GitHub Actions runs the steps above on every push to `main`, tags the image, and pushes to `ghcr.io/formswrite/schoolrise-app`. Production deploys then `docker pull` the tag instead of building locally.

Tracking issue: [production-roadmap.md item #9](roadmap.md#item-9-cicd-image-build-on-ghcr).

Until that lands, every operator runs the steps above on the deployment server (or rsyncs an image they built on their laptop). It works, but it's the load-bearing fragility we want to eliminate.
