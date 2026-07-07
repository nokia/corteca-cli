#!/usr/bin/env bash
# Run the same checks as build.yml locally.
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
RESET='\033[0m'

pass() { echo -e "${GREEN}✓ $1${RESET}"; }
fail() { echo -e "${RED}✗ $1${RESET}"; exit 1; }
warn() { echo -e "${YELLOW}⚠ $1${RESET}"; }
header() { echo -e "\n${BOLD}=== $1 ===${RESET}"; }

# ── Build ────────────────────────────────────────────────────────────────────
header "Build"
make || fail "build failed"
pass "build"

# ── Tests ────────────────────────────────────────────────────────────────────
header "Tests"
make test || fail "tests failed"
pass "tests"

# ── Go lint ──────────────────────────────────────────────────────────────────
header "Go lint (golangci-lint)"
if ! command -v golangci-lint &>/dev/null; then
    fail "golangci-lint not found — install it or open the devcontainer"
fi
golangci-lint run ./... || fail "Go lint failed"
pass "Go lint"

# ── YAML lint ────────────────────────────────────────────────────────────────
header "YAML lint (yamllint)"
if ! command -v yamllint &>/dev/null; then
    echo "yamllint not found, installing via apt..."
    sudo apt-get install -y -qq yamllint
fi
yamllint -c .yamllint.yaml . || fail "YAML lint failed"
pass "YAML lint"

# ── Markdown lint ────────────────────────────────────────────────────────────
header "Markdown lint (markdownlint-cli2)"
if ! command -v npx &>/dev/null; then
    warn "npx / Node.js not found — skipping markdown lint (runs in CI via GitHub Actions)"
else
    npx --yes markdownlint-cli2 --config .markdownlint.yaml "**/*.md" || fail "Markdown lint failed"
    pass "Markdown lint"
fi

# ─────────────────────────────────────────────────────────────────────────────
echo -e "\n${GREEN}${BOLD}All available checks passed.${RESET}"
