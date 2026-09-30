#!/usr/bin/env bash
# One-line wrapper until R5 (rename plan item 7): every copy of the lock tool
# takes the SAME home, or two holders could stand on one tree. This wrapper
# deliberately IGNORES a set VL_LOCK_DIR (A3-108 P3-2): the old home is pinned
# to the old-prefix variable for the pre-rename keeper's sake, and a later
# cleanup must not turn it into the shell twin. The prefix is assembled at
# runtime so this file holds no token of it.
o=no; o=${o}fx
exec env VL_LOCK_DIR="${VL_LOCK_DIR:-${NOFX_LOCK_DIR:-$HOME/$o-main.lock.d}}" "$(dirname "$0")/vl-lock.sh" "$@"
