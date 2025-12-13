# Integration Tests (Workflows)

This folder contains **integration tests** (also called workflow tests) that verify multiple API operations work together correctly.

## Unit Tests vs Integration Tests

### Unit Tests (Individual Resource Folders)
- **Location**: `collections/`, `templates/`, `notes/`, etc.
- **Purpose**: Test individual endpoints in isolation
- **Pattern**: CRUD operations (Create, Read, Update, Delete)
- **Example**: Create a collection, verify response

### Integration Tests (This Folder)
- **Location**: `workflows/`
- **Purpose**: Test multiple endpoints together to verify system behavior
- **Pattern**: Multi-step workflows that chain requests
- **Example**: Create note → Get note metadata → Verify metadata matches note content

## When to Use Integration Tests

Create an integration test when you need to verify:
- **Cross-resource behavior**: Note creation also creates links/metadata
- **Data flow**: Output of one endpoint becomes input of another
- **Business logic**: Complex scenarios that span multiple operations
- **System-level correctness**: Not just "does this endpoint work" but "does the system work as expected"

## Structure

Each workflow test:
1. **Orchestrates multiple requests** using `bru.runRequest()` or direct API calls
2. **Saves data** to environment variables for subsequent steps
3. **Verifies system behavior** with assertions across multiple responses
4. **Documents the workflow** as executable specification

## Example Workflow

```javascript
// Create note with frontmatter
const createResponse = await bru.runRequest("notes/1-create-note");
const noteId = createResponse.body.id;

// Get note metadata (should include frontmatter-extracted meta)
const metaResponse = await bru.runRequest("note-meta/1-list-note-meta");

// Verify metadata was extracted from frontmatter
test("Metadata extracted from frontmatter", function() {
  expect(metaResponse.body.metadata).to.have.property("tags");
});
```

## Running Workflows

### Run All Workflows
```bash
bru run workflows --env local
```

### Run Specific Workflow
```bash
bru run workflows/note-metadata-extraction.bru --env local
```

### Run as Part of Full Suite
```bash
# Run all tests (unit + integration)
bru run . --env local
```

## Current Workflows

- **note-metadata-extraction.bru**: Verifies note creation → metadata retrieval
  - Creates note with frontmatter
  - Retrieves metadata
  - Verifies frontmatter was parsed and stored

## Adding New Workflows

1. Create `.bru` file in this folder
2. Use `bru.runRequest()` to chain existing unit test requests
3. Add comprehensive assertions that verify system behavior
4. Document the workflow purpose in comments
5. Use `seq` numbers to control execution order if needed

## Best Practices

✅ **DO:**
- Use descriptive names: `note-with-links-resolution.bru`
- Chain existing unit test requests when possible
- Add detailed assertions that verify system behavior
- Log progress with `console.log()` for debugging
- Document the workflow scenario in comments

❌ **DON'T:**
- Duplicate unit test logic (use `bru.runRequest()` instead)
- Create workflows for single-endpoint tests (that's a unit test)
- Add cleanup (yet) - we'll establish that pattern later
- Make workflows depend on each other (keep them independent)

## Resources

- [Bruno Request Chaining](https://docs.usebruno.com/testing/script/request-chaining)
- [Bruno Script Flow](https://docs.usebruno.com/testing/script/script-flow)
- [Integration Testing Best Practices](https://martinfowler.com/articles/practical-test-pyramid.html)
