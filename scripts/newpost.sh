#!/bin/sh
TITLE="${1:-New Post}"
SLUG=$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$//')
YEAR=$(date +%Y)
DIR="content/posts/$YEAR"
mkdir -p "$DIR"
FILE="$DIR/$SLUG.md"

cat > "$FILE" <<EOF
---
title: $TITLE
date: $(date +%Y-%m-%d)
published: true
description:
---

EOF

echo "Created $FILE"
