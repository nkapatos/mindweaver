#!/usr/bin/env bash
# Generate Lua type annotations from protobuf definitions
# Usage: ./generate-types.sh <proto_dir> <output_file>

set -euo pipefail

PROTO_DIR="${1:?Usage: $0 <proto_dir> <output_file>}"
OUTPUT_FILE="${2:?Usage: $0 <proto_dir> <output_file>}"

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Start the output file
cat > "$OUTPUT_FILE" << 'HEADER'
-- AUTO-GENERATED from pkg/proto/mind/v3/*.proto
-- Do not edit manually. Run: task neoweaver:types:generate

local M = {}

HEADER

# Map protobuf types to Lua types
map_type() {
    local proto_type="$1"
    case "$proto_type" in
        string) echo "string" ;;
        int32|int64|uint32|uint64|sint32|sint64|fixed32|fixed64|sfixed32|sfixed64)
            echo "integer" ;;
        float|double) echo "number" ;;
        bool) echo "boolean" ;;
        bytes) echo "string" ;;
        "google.protobuf.Timestamp") echo "string" ;;
        "google.protobuf.Empty") echo "nil" ;;
        "map<"*) echo "table<string, string>" ;;
        *) echo "$proto_type" ;;
    esac
}

# Process each proto file
for proto_file in "$PROTO_DIR"/*.proto; do
    [ -f "$proto_file" ] || continue
    
    filename=$(basename "$proto_file" .proto)
    echo "-- From $filename.proto" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    current_message=""
    in_message=false
    
    while IFS= read -r line; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*// ]] && continue
        
        # Detect message start
        if [[ "$line" =~ ^[[:space:]]*message[[:space:]]+([A-Za-z_][A-Za-z0-9_]*) ]]; then
            current_message="${BASH_REMATCH[1]}"
            in_message=true
            echo "---@class mind.v3.$current_message" >> "$OUTPUT_FILE"
            continue
        fi
        
        # Detect message end (simple brace counting - assumes well-formed protos)
        if [[ "$in_message" == true && "$line" =~ ^[[:space:]]*\}[[:space:]]*$ ]]; then
            in_message=false
            echo "" >> "$OUTPUT_FILE"
            continue
        fi
        
        # Parse field definitions inside message
        if [[ "$in_message" == true ]]; then
            # Handle: optional type name = N;
            # Handle: type name = N;
            # Handle: repeated type name = N;
            # Handle: map<key, value> name = N;
            
            # Extract comment if present
            comment=""
            if [[ "$line" =~ //[[:space:]]*(.*) ]]; then
                comment="${BASH_REMATCH[1]}"
            fi
            
            # Check for optional
            is_optional=false
            if [[ "$line" =~ ^[[:space:]]*(optional)[[:space:]] ]]; then
                is_optional=true
            fi
            
            # Check for repeated
            is_repeated=false
            if [[ "$line" =~ ^[[:space:]]*(repeated)[[:space:]] ]]; then
                is_repeated=true
            fi
            
            # Parse field: match type and name
            # Patterns: optional? repeated? type name = N [(annotations)]
            # Strip validation annotations first for cleaner parsing
            clean_line="${line%%\[*}"
            if [[ "$clean_line" =~ ^[[:space:]]*(optional[[:space:]]+|repeated[[:space:]]+)?(map\<[^>]+\>|[A-Za-z_.]+)[[:space:]]+([a-z_][a-z0-9_]*)[[:space:]]*= ]]; then
                field_type="${BASH_REMATCH[2]}"
                field_name="${BASH_REMATCH[3]}"
                
                # Map the type
                lua_type=$(map_type "$field_type")
                
                # Handle repeated as array
                if [[ "$is_repeated" == true ]]; then
                    lua_type="${lua_type}[]"
                fi
                
                # Build the annotation
                if [[ "$is_optional" == true ]]; then
                    annotation="---@field ${field_name}? ${lua_type}"
                else
                    annotation="---@field ${field_name} ${lua_type}"
                fi
                
                # Add comment if present
                if [[ -n "$comment" ]]; then
                    annotation="$annotation $comment"
                fi
                
                echo "$annotation" >> "$OUTPUT_FILE"
            fi
        fi
    done < "$proto_file"
done

# Close the module
cat >> "$OUTPUT_FILE" << 'FOOTER'
return M
FOOTER

echo "Generated: $OUTPUT_FILE"
