#!/bin/sh
find public/images -type f -iname "*.heic" | while read heic; do
  jpg="${heic%.*}.jpg"
  if [ ! -f "$jpg" ] || [ "$heic" -nt "$jpg" ] 2>/dev/null || [ ! -s "$jpg" ]; then
    sips -Z 1600 -s format jpeg "$heic" --out "$jpg" > /dev/null 2>&1
    rm "$heic"
    echo "Converted: $heic"
  fi
done
