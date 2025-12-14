#!/usr/bin/env bash
# Convert TypeScript protobuf types to Lua type annotations
# Usage: ./ts-to-lua.sh <ts_gen_dir> <output_file>

set -euo pipefail

TS_GEN_DIR="${1:?Usage: $0 <ts_gen_dir> <output_file>}"
OUTPUT_FILE="${2:?Usage: $0 <ts_gen_dir> <output_file>}"

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Start the output file
cat > "$OUTPUT_FILE" << 'HEADER'
-- AUTO-GENERATED from proto/mind/v3/*.proto via TypeScript
-- Do not edit manually. Run: task neoweaver:types:generate

local M = {}

HEADER

# Map TypeScript types to Lua types
map_ts_to_lua() {
    local ts_type="$1"
    case "$ts_type" in
        string) echo "string" ;;
        number|bigint) echo "integer" ;;
        boolean) echo "boolean" ;;
        "Timestamp") echo "string" ;;
        "Date") echo "string" ;;
        *"[]") 
            # Array type - extract inner type and make it an array
            inner="${ts_type%\[\]}"
            echo "$(map_ts_to_lua "$inner")[]"
            ;;
        "Record<"*) echo "table<string, any>" ;;
        *) echo "$ts_type" ;;
    esac
}

# Process each TypeScript file
for ts_file in "$TS_GEN_DIR"/mind/v3/*.ts; do
    [ -f "$ts_file" ] || continue
    
    filename=$(basename "$ts_file" .ts)
    echo "-- From $filename.proto" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    # Extract interface/type definitions
    while IFS= read -r line; do
        # Match: export interface ClassName {
        if [[ "$line" =~ ^export[[:space:]]+(interface|type)[[:space:]]+([A-Za-z_][A-Za-z0-9_]*) ]]; then
            class_name="${BASH_REMATCH[2]}"
            echo "---@class mind.v3.$class_name" >> "$OUTPUT_FILE"
            continue
        fi
        
        # Match field definitions: fieldName?: type;
        # Or: fieldName: type;
        if [[ "$line" =~ ^[[:space:]]+([a-z_][a-zA-Z0-9_]*)\??:[[:space:]]*([^;]+); ]]; then
            field_name="${BASH_REMATCH[1]}"
            field_type="${BASH_REMATCH[2]}"
            
            # Check if optional (has ? before :)
            is_optional=false
            if [[ "$line" =~ ^[[:space:]]+[a-z_][a-zA-Z0-9_]*\? ]]; then
                is_optional=true
            fi
            
            # Clean up the type (remove comments, whitespace)
            field_type=$(echo "$field_type" | sed 's/\/\/.*//' | xargs)
            
            # Map TypeScript type to Lua type
            lua_type=$(map_ts_to_lua "$field_type")
            
            # Build annotation
            if [[ "$is_optional" == true ]]; then
                echo "---@field ${field_name}? ${lua_type}" >> "$OUTPUT_FILE"
            else
                echo "---@field ${field_name} ${lua_type}" >> "$OUTPUT_FILE"
            fi
        fi
        
        # Detect end of interface/type (closing brace)
        if [[ "$line" =~ ^[[:space:]]*\}[[:space:]]*$ ]]; then
            echo "" >> "$OUTPUT_FILE"
        fi
    done < "$ts_file"
done

# Close the module
cat >> "$OUTPUT_FILE" << 'FOOTER'
return M
FOOTER

echo "Generated Lua types: $OUTPUT_FILE"
