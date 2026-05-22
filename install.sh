#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
OC_CONFIG_DIR="${OPENCODE_CONFIG_DIR:-$HOME/.config/opencode}"
MANIFEST="$OC_CONFIG_DIR/.ntech-team-kit-manifest"
MODE="copy"
DRY_RUN=0
COMMAND="install"

SKILLS=(
  check-compiler-errors
  control-cli
  control-ui
  deslop
  fix-ci
  fix-merge-conflicts
  get-pr-comments
  loop-on-ci
  make-pr-easy-to-review
  new-branch-and-pr
  pr-review-canvas
  review-and-ship
  run-smoke-tests
  thermo-nuclear-code-quality-review
  verify-this
  weekly-review
  what-did-i-get-done
  workflow-from-chats
)

AGENTS=(
  ci-watcher
  thermo-nuclear-code-quality-review
)

COMMANDS=(
  review-and-ship
  loop-on-ci
  verify-this
  run-smoke-tests
  fix-ci
  new-branch-and-pr
  make-pr-easy-to-review
  fix-merge-conflicts
)

RULES=(
  no-inline-imports
  typescript-exhaustive-switch
)

PLUGIN_DEP="@opencode-ai/plugin"
PLUGIN_DEP_VERSION="^1.14.0"

usage() {
  cat <<'EOF'
ntech-team-kit installer

Usage: install.sh [OPTIONS] [COMMAND]

Commands:
  install     Install skills, agents, commands, rules, and plugins (default)
  uninstall   Remove all installed files
  status      Show what is currently installed

Options:
  --copy      Copy files (default)
  --link      Symlink files (not recommended — OpenCode does not follow symlinks)
  --dry-run   Show what would be done without doing it
  -h, --help  Show this help
EOF
}

log() { echo "[ntech-team-kit] $*"; }
dry_log() { echo "[ntech-team-kit] [dry-run] $*"; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --copy) MODE="copy"; shift ;;
    --link) MODE="link"; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    install|uninstall|status) COMMAND="$1"; shift ;;
    *) echo "Unknown argument: $1"; usage; exit 1 ;;
  esac
done

manifest_add() {
  echo "$1" >> "$MANIFEST"
}

manifest_remove() {
  if [[ -f "$MANIFEST" ]]; then
    local tmp
    tmp=$(mktemp)
    grep -v "^$1\$" "$MANIFEST" > "$tmp" || true
    mv "$tmp" "$MANIFEST"
  fi
}

ensure_plugin_dependency() {
  local pkg="$OC_CONFIG_DIR/package.json"
  if [[ $DRY_RUN -eq 1 ]]; then
    dry_log "ensure $PLUGIN_DEP@$PLUGIN_DEP_VERSION in $pkg"
    return
  fi

  if [[ ! -f "$pkg" ]]; then
    cat > "$pkg" <<EOF
{
  "dependencies": {
    "$PLUGIN_DEP": "$PLUGIN_DEP_VERSION"
  }
}
EOF
    log "created $pkg with $PLUGIN_DEP dependency"
    return
  fi

  if grep -q '"@opencode-ai/plugin"' "$pkg"; then
    log "$pkg already contains $PLUGIN_DEP"
    return
  fi

  if command -v node >/dev/null 2>&1; then
    local tmp
    tmp=$(mktemp)
    node -e '
const fs = require("fs")
const path = process.argv[1]
const pkg = JSON.parse(fs.readFileSync(path, "utf8"))
pkg.dependencies = pkg.dependencies || {}
pkg.dependencies["@opencode-ai/plugin"] = "^1.14.0"
process.stdout.write(JSON.stringify(pkg, null, 2) + "\n")
' "$pkg" > "$tmp"
    cp "$pkg" "$pkg.ntech-team-kit.bak"
    mv "$tmp" "$pkg"
    log "added $PLUGIN_DEP to $pkg (backup: $pkg.ntech-team-kit.bak)"
  else
    log "warning: node not found; add $PLUGIN_DEP to $pkg manually if the plugin fails to load"
  fi
}

do_install_file() {
  local src="$1" dest="$2"
  if [[ $DRY_RUN -eq 1 ]]; then
    dry_log "$MODE $src -> $dest"
    return
  fi

  # Ensure the destination's parent directory exists (defensive against any
  # prior partial state or brace-expansion quirks on some shells)
  mkdir -p "$(dirname "$dest")"

  if [[ ! -e "$src" ]]; then
    echo "[ntech-team-kit] ERROR: source file missing: $src"
    echo "[ntech-team-kit]        REPO_DIR appears to be: $REPO_DIR"
    echo "[ntech-team-kit]        This usually means the installed kit is incomplete."
    echo "[ntech-team-kit]        Try: NTECH_TEAM_KIT_ROOT=/path/to/clone ntech-team-kit install"
    exit 1
  fi

  if [[ "$MODE" == "link" ]]; then
    ln -sfn "$src" "$dest"
  else
    cp -f "$src" "$dest"
  fi
  manifest_add "$dest"
  log "$MODE $src -> $dest"
}

do_remove_file() {
  local dest="$1"
  if [[ -L "$dest" ]]; then
    if [[ $DRY_RUN -eq 1 ]]; then
      dry_log "rm $dest"
    else
      rm -f "$dest"
      log "removed $dest"
    fi
  elif [[ -f "$dest" ]]; then
    if [[ $DRY_RUN -eq 1 ]]; then
      dry_log "rm $dest"
    else
      rm -f "$dest"
      log "removed $dest"
    fi
  fi
  manifest_remove "$dest"
}

do_install() {
  # Create top-level directories explicitly (brace expansion after a variable
  # can be unreliable on some macOS shells / quoting contexts)
  mkdir -p "$OC_CONFIG_DIR/skills" \
           "$OC_CONFIG_DIR/agents" \
           "$OC_CONFIG_DIR/commands" \
           "$OC_CONFIG_DIR/rules" \
           "$OC_CONFIG_DIR/plugins"
  : > "$MANIFEST"

  for skill in "${SKILLS[@]}"; do
    mkdir -p "$OC_CONFIG_DIR/skills/$skill"
    do_install_file "$REPO_DIR/skills/$skill/SKILL.md" "$OC_CONFIG_DIR/skills/$skill/SKILL.md"
    if [[ "$skill" == "pr-review-canvas" ]]; then
      for asset in renderer.js styles.css template.html; do
        do_install_file "$REPO_DIR/skills/$skill/$asset" "$OC_CONFIG_DIR/skills/$skill/$asset"
      done
    fi
  done

  for agent in "${AGENTS[@]}"; do
    do_install_file "$REPO_DIR/agents/$agent.md" "$OC_CONFIG_DIR/agents/$agent.md"
  done

  for cmd in "${COMMANDS[@]}"; do
    do_install_file "$REPO_DIR/commands/$cmd.md" "$OC_CONFIG_DIR/commands/$cmd.md"
  done

  for rule in "${RULES[@]}"; do
    do_install_file "$REPO_DIR/rules/$rule.md" "$OC_CONFIG_DIR/rules/$rule.md"
  done

  rm -f "$OC_CONFIG_DIR/plugins/package.json"
  do_install_file "$REPO_DIR/plugins/ci-watcher.ts" "$OC_CONFIG_DIR/plugins/ci-watcher.ts"
  ensure_plugin_dependency

  if [[ $DRY_RUN -eq 0 ]]; then
    log "install complete ($MODE mode)"
    if [[ "$MODE" == "link" ]]; then
      log "  warning: symlink mode — OpenCode may not discover symlinked skills/agents/commands"
    fi
    log "  skills:   ${#SKILLS[@]}"
    log "  agents:   ${#AGENTS[@]}"
    log "  commands: ${#COMMANDS[@]}"
    log "  rules:    ${#RULES[@]}"
    log "  plugins:  1 (ci-watcher)"
    log ""
    log "To enable background CI watching, set:"
    log "  export OPENCODE_NTECH_CI_WATCH=1"
    log ""
    log "To merge rules into AGENTS.md, see opencode.jsonc instructions glob."
    log "Manifest written to $MANIFEST"
  fi
}

do_uninstall() {
  if [[ ! -f "$MANIFEST" ]]; then
    log "no manifest found at $MANIFEST — nothing to uninstall"
    return
  fi
  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    do_remove_file "$path"
  done < "$MANIFEST"

  for skill in "${SKILLS[@]}"; do
    local dir="$OC_CONFIG_DIR/skills/$skill"
    if [[ -d "$dir" ]] && [[ -z "$(ls -A "$dir" 2>/dev/null)" ]]; then
      if [[ $DRY_RUN -eq 1 ]]; then
        dry_log "rmdir $dir"
      else
        rmdir "$dir" 2>/dev/null || true
      fi
    fi
  done

  if [[ $DRY_RUN -eq 0 ]]; then
    rm -f "$MANIFEST"
    log "uninstall complete"
  fi
}

do_status() {
  if [[ ! -f "$MANIFEST" ]]; then
    log "not installed (no manifest at $MANIFEST)"
    return
  fi
  local count
  count=$(wc -l < "$MANIFEST")
  log "$count files tracked in manifest"
  local broken=0
  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    if [[ ! -e "$path" ]]; then
      log "  MISSING: $path"
      broken=$((broken + 1))
    fi
  done < "$MANIFEST"
  if [[ $broken -eq 0 ]]; then
    log "all files present"
  else
    log "$broken files missing — consider reinstalling"
  fi
}

case "$COMMAND" in
  install) do_install ;;
  uninstall) do_uninstall ;;
  status) do_status ;;
esac
