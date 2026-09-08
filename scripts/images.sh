#!/bin/sh
set -eu
command -v sips >/dev/null 2>&1 || { echo "sips is required to convert HEIC photos" >&2; exit 1; }
# Run each conversion in its own process so paths containing spaces/newlines are safe.
find public/images -type f -iname '*.heic' -exec sh -eu -c '
 for heic do
  jpg="${heic%.*}.jpg"
  if [ ! -s "$jpg" ] || [ "$heic" -nt "$jpg" ]; then
   tmp=$(mktemp "${jpg}.XXXXXX")
   trap '\''rm -f "$tmp"'\'' EXIT
   if sips -Z 1600 -s format jpeg "$heic" --out "$tmp" >/dev/null 2>&1 && [ -s "$tmp" ] && sips -g format "$tmp" 2>/dev/null | grep -q "format: jpeg"; then
    mv "$tmp" "$jpg"
    rm "$heic"
    echo "Converted: $heic"
   else
    echo "Conversion failed; original retained: $heic" >&2
    exit 1
   fi
   trap - EXIT
  fi
 done
' sh {} +
