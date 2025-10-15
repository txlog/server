#!/bin/bash

# This script generates a single schema.sql file from all the .up.sql migration files.

OUTPUT_FILE="schema.sql"
MIGRATIONS_DIR="database/migrations"

# Clear the output file
> "$OUTPUT_FILE"

# Find all .up.sql files, sort them, and process each one
for file in $(ls -1 "$MIGRATIONS_DIR"/*.up.sql | sort -n); do
    echo "-- From file: $file" >> "$OUTPUT_FILE"
    cat "$file" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
done

echo "Schema generated in $OUTPUT_FILE"
