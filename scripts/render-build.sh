#!/usr/bin/env bash
# Render build entrypoint (referenced by render.yaml and usable in the
# dashboard's build command). Builds the React frontend first so the Go server
# can serve it from web/dist, then compiles the Go binary.
set -euo pipefail

# 1. Frontend: install + build -> web/dist (tsc --noEmit && vite build)
echo "==> Building frontend (web/dist)"
npm --prefix web ci
npm --prefix web run build

# 2. Go server: single binary named `app` (matches startCommand).
echo "==> Building Go server"
go build -tags netgo -ldflags '-s -w' -o app ./cmd/kweeks-server

echo "==> Build complete"
