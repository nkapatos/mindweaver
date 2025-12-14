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
        number) echo "number" ;;
        bigint) echo "integer" ;;
        boolean) echo "boolean" ;;
        "Timestamp") echo "string" ;;
        "Date") echo "string" ;;
        "{ [key: string]: string }") echo "table<string, string>" ;;
        *"[]") 
            # Array type - extract inner type and make it an array
            inner="${ts_type%\[\]}"
            echo "$(map_ts_to_lua "$inner")[]"
            ;;
        "Record<"*) echo "table<string, any>" ;;
        *) echo "$ts_type" ;;
    esac
}

# Process each TypeScript _pb.ts file (these contain message definitions)
for ts_file in "$TS_GEN_DIR"/v3/*_pb.ts; do
    [ -f "$ts_file" ] || continue
    
    filename=$(basename "$ts_file" _pb.ts)
    echo "-- From $filename.proto" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    in_type=false
    current_type=""
    
    while IFS= read -r line; do
        # Match: export type TypeName = Message<"mind.v3.TypeName"> & {
        if [[ "$line" =~ ^export[[:space:]]+type[[:space:]]+([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=[[:space:]]*Message\<\"mind\.v3\.([A-Za-z_][A-Za-z0-9_]*)\"\>[[:space:]]*\&[[:space:]]*\{ ]]; then
            current_type="${BASH_REMATCH[1]}"
            in_type=true
            echo "---@class mind.v3.$current_type" >> "$OUTPUT_FILE"
            continue
        fi
        
        # Detect end of type (closing brace followed by semicolon)
        if [[ "$in_type" == true && "$line" =~ ^[[:space:]]*\}\;[[:space:]]*$ ]]; then
            in_type=false
            echo "" >> "$OUTPUT_FILE"
            continue
        fi
        
        # Parse field definitions inside type
        if [[ "$in_type" == true ]]; then
            # Skip comment-only lines
            if [[ "$line" =~ ^[[:space:]]*\*.*$ ]] || [[ "$line" =~ ^[[:space:]]*\/\*.*$ ]] || [[ "$line" =~ ^[[:space:]]*\/\/.*$ ]]; then
                continue
            fi
            
            # Match field definition patterns:
            # fieldName: type;
            # fieldName?: type;
            # fieldName: { [key: string]: string };
            if [[ "$line" =~ ^[[:space:]]+([a-z][a-zA-Z0-9_]*)\??:[[:space:]]*([^\;]+)\;[[:space:]]*$ ]]; then
                field_name="${BASH_REMATCH[1]}"
                field_type="${BASH_REMATCH[2]}"
                
                # Check if optional (has ? after field name)
                is_optional=false
                if [[ "$line" =~ ^[[:space:]]+[a-z][a-zA-Z0-9_]*\? ]]; then
                    is_optional=true
                fi
                
                # Clean up the type (remove extra whitespace)
                field_type=$(echo "$field_type" | xargs)
                
                # Map TypeScript type to Lua type
                lua_type=$(map_ts_to_lua "$field_type")
                
                # Build annotation
                if [[ "$is_optional" == true ]]; then
                    echo "---@field ${field_name}? ${lua_type}" >> "$OUTPUT_FILE"
                else
                    echo "---@field ${field_name} ${lua_type}" >> "$OUTPUT_FILE"
                fi
            fi
        fi
    done < "$ts_file"
done

# Close the module
cat >> "$OUTPUT_FILE" << 'FOOTER'
return M
FOOTER

echo "Generated Lua types: $OUTPUT_FILE"
