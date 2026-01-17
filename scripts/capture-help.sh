#!/bin/bash
set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <manager>"
    echo "Example: $0 npm"
    exit 1
fi

MANAGER="$1"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REFS_DIR="$SCRIPT_DIR/../references/$MANAGER"

mkdir -p "$REFS_DIR"

case "$MANAGER" in
    npm)
        npm --version > "$REFS_DIR/version.txt"
        npm --help > "$REFS_DIR/help.txt" 2>&1 || true
        npm install --help > "$REFS_DIR/install.txt" 2>&1 || true
        npm uninstall --help > "$REFS_DIR/uninstall.txt" 2>&1 || true
        npm update --help > "$REFS_DIR/update.txt" 2>&1 || true
        npm list --help > "$REFS_DIR/list.txt" 2>&1 || true
        npm outdated --help > "$REFS_DIR/outdated.txt" 2>&1 || true
        npm audit --help > "$REFS_DIR/audit.txt" 2>&1 || true
        man npm-install > "$REFS_DIR/man-install.txt" 2>/dev/null || true
        ;;
    yarn)
        yarn --version > "$REFS_DIR/version.txt"
        yarn --help > "$REFS_DIR/help.txt" 2>&1 || true
        yarn add --help > "$REFS_DIR/add.txt" 2>&1 || true
        yarn remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        yarn install --help > "$REFS_DIR/install.txt" 2>&1 || true
        yarn upgrade --help > "$REFS_DIR/upgrade.txt" 2>&1 || true
        yarn list --help > "$REFS_DIR/list.txt" 2>&1 || true
        yarn outdated --help > "$REFS_DIR/outdated.txt" 2>&1 || true
        ;;
    pnpm)
        pnpm --version > "$REFS_DIR/version.txt"
        pnpm --help > "$REFS_DIR/help.txt" 2>&1 || true
        pnpm add --help > "$REFS_DIR/add.txt" 2>&1 || true
        pnpm remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        pnpm install --help > "$REFS_DIR/install.txt" 2>&1 || true
        pnpm update --help > "$REFS_DIR/update.txt" 2>&1 || true
        pnpm list --help > "$REFS_DIR/list.txt" 2>&1 || true
        pnpm outdated --help > "$REFS_DIR/outdated.txt" 2>&1 || true
        ;;
    bun)
        bun --version > "$REFS_DIR/version.txt"
        bun --help > "$REFS_DIR/help.txt" 2>&1 || true
        bun add --help > "$REFS_DIR/add.txt" 2>&1 || true
        bun remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        bun install --help > "$REFS_DIR/install.txt" 2>&1 || true
        bun update --help > "$REFS_DIR/update.txt" 2>&1 || true
        ;;
    bundler)
        bundle --version > "$REFS_DIR/version.txt"
        bundle help > "$REFS_DIR/help.txt" 2>&1 || true
        bundle help install > "$REFS_DIR/install.txt" 2>&1 || true
        bundle help add > "$REFS_DIR/add.txt" 2>&1 || true
        bundle help remove > "$REFS_DIR/remove.txt" 2>&1 || true
        bundle help update > "$REFS_DIR/update.txt" 2>&1 || true
        bundle help list > "$REFS_DIR/list.txt" 2>&1 || true
        bundle help outdated > "$REFS_DIR/outdated.txt" 2>&1 || true
        ;;
    cargo)
        cargo --version > "$REFS_DIR/version.txt"
        cargo --help > "$REFS_DIR/help.txt" 2>&1 || true
        cargo add --help > "$REFS_DIR/add.txt" 2>&1 || true
        cargo remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        cargo install --help > "$REFS_DIR/install.txt" 2>&1 || true
        cargo update --help > "$REFS_DIR/update.txt" 2>&1 || true
        cargo build --help > "$REFS_DIR/build.txt" 2>&1 || true
        ;;
    go)
        go version > "$REFS_DIR/version.txt"
        go help > "$REFS_DIR/help.txt" 2>&1 || true
        go help get > "$REFS_DIR/get.txt" 2>&1 || true
        go help install > "$REFS_DIR/install.txt" 2>&1 || true
        go help mod > "$REFS_DIR/mod.txt" 2>&1 || true
        go help mod tidy > "$REFS_DIR/mod-tidy.txt" 2>&1 || true
        go help mod download > "$REFS_DIR/mod-download.txt" 2>&1 || true
        go help list > "$REFS_DIR/list.txt" 2>&1 || true
        ;;
    uv)
        uv --version > "$REFS_DIR/version.txt"
        uv --help > "$REFS_DIR/help.txt" 2>&1 || true
        uv add --help > "$REFS_DIR/add.txt" 2>&1 || true
        uv remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        uv sync --help > "$REFS_DIR/sync.txt" 2>&1 || true
        uv lock --help > "$REFS_DIR/lock.txt" 2>&1 || true
        uv pip --help > "$REFS_DIR/pip.txt" 2>&1 || true
        ;;
    poetry)
        poetry --version > "$REFS_DIR/version.txt"
        poetry --help > "$REFS_DIR/help.txt" 2>&1 || true
        poetry add --help > "$REFS_DIR/add.txt" 2>&1 || true
        poetry remove --help > "$REFS_DIR/remove.txt" 2>&1 || true
        poetry install --help > "$REFS_DIR/install.txt" 2>&1 || true
        poetry update --help > "$REFS_DIR/update.txt" 2>&1 || true
        poetry show --help > "$REFS_DIR/show.txt" 2>&1 || true
        poetry lock --help > "$REFS_DIR/lock.txt" 2>&1 || true
        ;;
    *)
        echo "Unknown manager: $MANAGER"
        echo "Supported: npm, yarn, pnpm, bun, bundler, cargo, go, uv, poetry"
        exit 1
        ;;
esac

echo "Captured help for $MANAGER in $REFS_DIR"
ls -la "$REFS_DIR"
