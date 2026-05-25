#!/bin/bash
# Auto-commit and push on session stop — only if there are uncommitted changes
set -e

ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo '.')
cd "$ROOT"

if ! git rev-parse --git-dir > /dev/null 2>&1; then
  exit 0
fi

if [ -z "$(git status --porcelain)" ]; then
  exit 0
fi

COUNT=$(git status --porcelain | wc -l | tr -d ' ')
git add -A

git commit -m "auto: ${COUNT} file(s) changed

Generated with [Claude Code](https://claude.ai/code)
via [Happy](https://happy.engineering)

Co-Authored-By: Neil <neil@loggraph.local>" 2>&1 || exit 0

git push 2>&1 || exit 0
