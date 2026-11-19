#!/bin/sh
set -eu
umask 077
cd /opt/laytech-ai/CodeCrashReview
backup_dir=/home/laytech/.local/share/codecrashreview/backups
mkdir -p "$backup_dir"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
target="$backup_dir/database-$stamp.dump"
trap 'rm -f "$target.partial"' EXIT
docker compose -f compose.host.yml exec -T postgres pg_dump -U codecrash -d codecrash -Fc > "$target.partial"
docker compose -f compose.host.yml exec -T postgres pg_restore --list < "$target.partial" > /dev/null
mv "$target.partial" "$target"
sha256sum "$target" > "$target.sha256"
printf '%s
' "$target"
