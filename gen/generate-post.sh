#!/bin/bash

echo "POST SQLC!!"

rm gen.*.sql.go
for f in gen/*.sql.go; do
    sed -i 's/Queries/DBStore/gI' $f
    mv "$f" "gen.${f#gen/}"
done
rm -r gen