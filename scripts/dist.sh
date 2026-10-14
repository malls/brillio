#!/bin/sh
# make dist (docs/SPEC.md section 12): write the release zip
# dist/brillio-listings-YYYYMMDD.zip from the committed tree at HEAD. It is a
# git archive, so neither uncommitted changes nor the host's global gitignore
# (core.excludesFile) change what ships, and .git never does. The zip holds one
# top-level folder, brillio-listings-YYYYMMDD/. Check it from a fresh directory
# with scripts/verify-dist.sh <zip>.
#
# POSIX sh. Needs git and unzip.
set -eu

# What does not ship (SPEC section 12), one git pathspec per line. Only the
# first three name tracked files; the rest are belt and braces (none of those
# paths is tracked). To stop shipping a tracked file, add a line here, such as
# ':(exclude)CLAUDE.md' \, and say so in SPEC section 12.
set -- \
  ':(exclude).lattice' \
  ':(exclude)Coding_Test_Question_Paper.pdf' \
  ':(exclude)Instructions.pdf' \
  ':(exclude)dist' \
  ':(exclude,glob)**/node_modules/**' \
  ':(exclude,glob)**/*.db' \
  ':(exclude,glob)**/*.db-*' \
  ':(exclude,glob)**/.DS_Store' \
  ':(exclude,glob)**/._*' \
  ':(exclude,glob)**/__MACOSX/**'

die() {
  printf 'dist: %s\n' "$*" >&2
  exit 1
}

cd "$(dirname "$0")/.."

command -v git >/dev/null 2>&1 || die 'git is not installed'
command -v unzip >/dev/null 2>&1 || die 'unzip is not installed'
# Inside an extracted zip there is no repository (or someone else's): refuse.
top=$(git rev-parse --show-toplevel 2>/dev/null) || top=
[ "$top" = "$(pwd -P)" ] || die 'make dist builds from a git checkout of this repository'

name=brillio-listings-$(date +%Y%m%d)
out=dist/$name.zip
tmp=dist/.$name.zip.tmp
head=$(git log -1 --format='%h %s' HEAD)

# A dirty tree only warns: the zip is a pure function of HEAD either way. The
# board under .lattice/ moves all the time and never ships, so it is left out.
dirty=$(git -c core.excludesFile=/dev/null status --porcelain -- . ':(exclude).lattice')
if [ -n "$dirty" ]; then
  printf 'dist: warning: the zip is HEAD %s; these uncommitted changes are NOT in it:\n' "${head%% *}" >&2
  printf '%s\n' "$dirty" | sed -n '1,20s/^/dist:   /p' >&2
  n=$(printf '%s\n' "$dirty" | wc -l)
  [ "$n" -le 20 ] || printf 'dist:   ... and %d more\n' "$((n - 20))" >&2
fi

mkdir -p dist
# Written beside the target and renamed, so neither a failure nor an interrupt
# leaves a partial zip.
trap 'rm -f "$tmp"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
git archive --format=zip --prefix="$name/" -o "$tmp" HEAD -- . "$@" || die 'git archive failed'
mv -f "$tmp" "$out"

bytes=$(($(wc -c <"$out")))
files=$(unzip -Z1 "$out" | grep -v '/$' | sed "s|^$name/||")
count=$(printf '%s\n' "$files" | wc -l)
tracked=$(git ls-tree -r --name-only HEAD | wc -l)
# Files per top-level path, root files first: "8 root, 46 backend/, ...".
per_top=$(printf '%s\n' "$files" |
  awk -F/ '{ k = (NF > 1) ? $1 "/" : ""; n[k]++ } END { for (k in n) print k, n[k] }' |
  sort | awk '{ s = s (NR > 1 ? ", " : "") (NF > 1 ? $2 " " $1 : $1 " root") } END { print s }')

# Tracked files each plain exclusion drops: "366 .lattice/, 1 Instructions.pdf".
# git ls-tree takes no pathspec magic, so the globs (nothing tracked) are left
# to the total.
dropped=
for x in "$@"; do
  p=${x#':(exclude)'}
  [ "$p" != "$x" ] || continue
  n=$(($(git ls-tree -r --name-only HEAD -- "$p" | wc -l)))
  if [ -d "$p" ]; then p=$p/; fi
  [ "$n" -eq 0 ] || dropped="$dropped${dropped:+, }$n $p"
done

echo "dist: source   HEAD $head"
echo "dist: wrote    $out"
echo "dist: size     $bytes bytes ($(((bytes + 1023) / 1024)) KiB)"
echo "dist: files    $((count)) ($per_top)"
echo "dist: excluded $((tracked - count)) tracked files (${dropped:-none})"
echo "dist: check it with: scripts/verify-dist.sh $out"
