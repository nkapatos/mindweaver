#!/usr/bin/env bun
/**
 * Convert TypeScript protobuf types to Lua type annotations
 * Usage: bun run ts-to-lua.ts <ts_gen_dir> <output_file>
 */

import * as fs from "fs";
import * as path from "path";
import ts from "typescript";

// Parse command line arguments
const args = process.argv.slice(2);
if (args.length !== 2) {
  console.error("Usage: bun run ts-to-lua.ts <ts_gen_dir> <output_file>");
  process.exit(1);
}

const [tsGenDir, outputFile] = args;

// Type mapping from TypeScript to Lua
function mapTypeToLua(tsType: string): string {
  // Handle arrays
  if (tsType.endsWith("[]")) {
    const innerType = tsType.slice(0, -2);
    return `${mapTypeToLua(innerType)}[]`;
  }

  // Handle specific types
  const typeMap: Record<string, string> = {
    string: "string",
    number: "number",
    bigint: "integer",
    boolean: "boolean",
    Timestamp: "string",
    Date: "string",
  };

  // Check if it's a map type
  if (tsType.includes("[key: string]")) {
    return "table<string, string>";
  }

  // Check if it's a Record type
  if (tsType.startsWith("Record<")) {
    return "table<string, any>";
  }

  return typeMap[tsType] || tsType;
}

// Parse a TypeScript file and extract type definitions
function parseTypeScriptFile(filePath: string): Array<{
  name: string;
  fields: Array<{ name: string; type: string; optional: boolean }>;
}> {
  const sourceCode = fs.readFileSync(filePath, "utf-8");
  const sourceFile = ts.createSourceFile(
    filePath,
    sourceCode,
    ts.ScriptTarget.Latest,
    true
  );

  const types: Array<{
    name: string;
    fields: Array<{ name: string; type: string; optional: boolean }>;
  }> = [];

  function visit(node: ts.Node) {
    // Look for: export type TypeName = Message<"mind.v3.TypeName"> & { ... }
    if (ts.isTypeAliasDeclaration(node)) {
      // Check if it has export modifier
      const hasExport = node.modifiers?.some(
        (m) => m.kind === ts.SyntaxKind.ExportKeyword
      );

      if (!hasExport) {
        ts.forEachChild(node, visit);
        return;
      }

      const typeName = node.name.text;

      // Check if it's a Message type intersection
      if (ts.isIntersectionTypeNode(node.type)) {
        const intersectionTypes = node.type.types;

        // Find the object type part (the { ... } with fields)
        const objectType = intersectionTypes.find((t) =>
          ts.isTypeLiteralNode(t)
        ) as ts.TypeLiteralNode | undefined;

        if (objectType) {
          const fields: Array<{
            name: string;
            type: string;
            optional: boolean;
          }> = [];

          for (const member of objectType.members) {
            if (ts.isPropertySignature(member) && member.name) {
              const fieldName = member.name.getText(sourceFile);
              const isOptional = !!member.questionToken;

              let fieldType = "any";
              if (member.type) {
                fieldType = member.type.getText(sourceFile);
              }

              fields.push({
                name: fieldName,
                type: mapTypeToLua(fieldType),
                optional: isOptional,
              });
            }
          }

          types.push({ name: typeName, fields });
        }
      }
    }

    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
  return types;
}

// Main logic
function main() {
  // Ensure output directory exists
  const outputDir = path.dirname(outputFile);
  fs.mkdirSync(outputDir, { recursive: true });

  // Start output with header
  let output = `--- Type definitions for MindWeaver API
---
--- AUTO-GENERATED from Protocol Buffer definitions (proto/mind/v3/*.proto).
--- Do not edit manually.
---
--- To regenerate these types:
---   task neoweaver:types:generate
---
--- This file provides LuaLS type annotations for the MindWeaver v3 API,
--- enabling autocomplete and type checking in Neovim plugin development.
---
---@module neoweaver.types
local M = {}

`;

  // Find all TypeScript files
  const v3Dir = path.join(tsGenDir, "v3");
  if (!fs.existsSync(v3Dir)) {
    console.error(`Directory not found: ${v3Dir}`);
    process.exit(1);
  }

  const tsFiles = fs
    .readdirSync(v3Dir)
    .filter((f) => f.endsWith("_pb.ts"))
    .sort();

  // Process each file
  for (const tsFile of tsFiles) {
    const filePath = path.join(v3Dir, tsFile);
    const protoName = tsFile.replace("_pb.ts", ".proto");

    output += `-- From ${protoName}\n\n`;

    const types = parseTypeScriptFile(filePath);

    for (const type of types) {
      output += `---@class mind.v3.${type.name}\n`;

      for (const field of type.fields) {
        const optionalMarker = field.optional ? "?" : "";
        output += `---@field ${field.name}${optionalMarker} ${field.type}\n`;
      }

      output += "\n";
    }
  }

  // Close the module
  output += "return M\n";

  // Write output file
  fs.writeFileSync(outputFile, output, "utf-8");

  console.log(`Generated Lua types: ${outputFile}`);
  console.log(
    `  - Processed ${tsFiles.length} TypeScript files from ${v3Dir}`
  );
}

main();
