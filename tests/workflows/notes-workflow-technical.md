# Notes Workflow - User Testing Guide

## Overview
This document describes the core workflows for managing notes in Mindweaver and the expected behaviors that users/client applications should verify.

---

## Workflow 1: Create a Simple Note

### User Actions
1. Create a new note with title and body
2. Verify the note was created successfully

### Expected Results
- **HTTP Status**: 201 Created
- **Response includes**:
  - `id` (integer)
  - `uuid` (UUID string)
  - `created_at` (ISO8601 timestamp)
  - `version` (starts at 1)
  - `etag` (hashed version for optimistic locking)
- **Response header**: `ETag` header is set

### Verification Steps
```
POST /api/mind/notes
{
  "title": "My First Note",
  "body": "This is the content of my note."
}

✓ Status is 201
✓ Response contains id, uuid, created_at, version, etag
✓ ETag header is present
✓ version equals 1 (first version)
```

---

## Workflow 2: Create a Note with Markdown & Frontmatter

### User Actions
1. Create a note with markdown body that includes YAML frontmatter
2. Frontmatter should be extracted and stored separately
3. WikiLinks in the body should be parsed and stored as links

### Expected Results
- **HTTP Status**: 201 Created
- **Frontmatter** is automatically extracted from the body and stored in the database
- **WikiLinks** (e.g., `[[Other Note]]`) are extracted and stored as note links
- If links reference non-existent notes, they are stored as pending links
- **Body** returned in GET requests includes the composed frontmatter

### Verification Steps
```
POST /api/mind/notes
{
  "title": "Book Notes",
  "body": "---\nauthor: James Clear\nrating: 5\n---\n\nThis book is about [[Atomic Habits]] and [[Behavior Change]]."
}

✓ Status is 201
✓ Note is created with extracted frontmatter
✓ GET /api/mind/notes/:id returns body WITH frontmatter composed
✓ GET /api/mind/notes/:id/links shows links to "Atomic Habits" and "Behavior Change"
✓ Links may be pending (dest_id is null) if target notes don't exist yet
```

---

## Workflow 3: Retrieve a Note

### User Actions
1. Get a note by ID
2. Optionally request related data (meta, tags, links)
3. Optionally select specific fields

### Expected Results
- **HTTP Status**: 200 OK
- **Response includes**: Full note data with all fields
- **Body field**: Always includes composed frontmatter (if frontmatter exists)
- **ETag header**: Set for version tracking
- **Field selection**: If `?fields=id,title` only those fields are returned
- **Relations**: If `?include=meta,tags,links` related data is included

### Verification Steps
```
GET /api/mind/notes/:id

✓ Status is 200
✓ Response contains complete note data
✓ Body includes composed frontmatter
✓ ETag header is present

GET /api/mind/notes/:id?fields=id,title

✓ Response only contains id and title fields

GET /api/mind/notes/:id?include=meta,tags,links

✓ Response includes nested meta, tags, and links arrays
```

---

## Workflow 4: Update a Note (Optimistic Locking)

### User Actions
1. Get a note to retrieve its current ETag
2. Update the note with If-Match header containing the ETag
3. Handle potential conflicts if note was modified by another client

### Expected Results
- **Successful Update**:
  - HTTP Status: 200 OK
  - New version number (incremented)
  - New ETag for next update
  
- **Missing If-Match Header**:
  - HTTP Status: 428 Precondition Required
  - Error message: "If-Match header is required for updates"
  
- **ETag Mismatch** (concurrent modification):
  - HTTP Status: 409 Conflict
  - Error message: "ETag mismatch - resource has been modified"

### Verification Steps
```
# Step 1: Get current note
GET /api/mind/notes/:id
Response: { ..., "version": 1, "etag": "W/\"abc123\"" }

# Step 2: Update with If-Match
PUT /api/mind/notes/:id
Headers: If-Match: W/"abc123"
{
  "title": "Updated Title"
}

✓ Status is 200
✓ Response version is 2 (incremented)
✓ Response etag is different from original
✓ New ETag header is set

# Step 3: Try update without If-Match
PUT /api/mind/notes/:id
{
  "title": "Another Update"
}

✓ Status is 428
✓ Error message about missing If-Match header

# Step 4: Try update with stale ETag
PUT /api/mind/notes/:id
Headers: If-Match: W/"old_etag"
{
  "title": "Stale Update"
}

✓ Status is 409
✓ Error message about ETag mismatch
```

---

## Workflow 5: Update Note Body (WikiLinks Parsing)

### User Actions
1. Update a note's body content
2. New WikiLinks should be parsed and stored
3. Old links should be removed

### Expected Results
- **HTTP Status**: 200 OK
- **Old links** are deleted
- **New links** from updated body are created
- **Backlinks** to this note are resolved if target notes now exist

### Verification Steps
```
# Original note has: "See [[Note A]]"
GET /api/mind/notes/:id/links
Response: [{ "target_title": "Note A", "dest_id": null }]  # Pending link

# Update body to reference different note
PUT /api/mind/notes/:id
Headers: If-Match: W/"current_etag"
{
  "body": "Now see [[Note B]] instead."
}

✓ Status is 200
✓ GET /api/mind/notes/:id/links shows only link to "Note B"
✓ Link to "Note A" is removed
```

---

## Workflow 6: List Notes with Filtering

### User Actions
1. List all notes
2. Filter by tag
3. Filter by metadata key
4. Combine filters

### Expected Results
- **HTTP Status**: 200 OK
- **Response structure**: `{ "data": { "kind": "note#list", "items": [...], "etag": "..." } }`
- **List ETag**: Computed from all items for cache validation
- **Filtering**: Only notes matching criteria are returned

### Verification Steps
```
GET /api/mind/notes

✓ Status is 200
✓ Returns array of all notes
✓ Response includes list etag
✓ ETag header is set

GET /api/mind/notes?tag_id=1

✓ Returns only notes tagged with tag_id=1

GET /api/mind/notes?meta_key=author

✓ Returns only notes that have 'author' metadata

GET /api/mind/notes?tag_id=1&meta_key=author

✓ Returns notes matching both criteria
```

---

## Workflow 7: Delete a Note

### User Actions
1. Delete a note by ID
2. Verify deletion was successful

### Expected Results
- **HTTP Status**: 200 OK
- **Response**: Confirmation with `{ "kind": "note", "id": X, "deleted": true }`
- **Side effects**: Associated links, metadata, and tags are cascade deleted

### Verification Steps
```
DELETE /api/mind/notes/:id

✓ Status is 200
✓ Response confirms deletion with deleted: true
✓ Subsequent GET returns 404
✓ Associated metadata is deleted
✓ Outgoing links are deleted
✓ Backlinks referencing this note are updated/deleted
```

---

## Workflow 8: Access Sub-Resources

### User Actions
1. Get note metadata: `/api/mind/notes/:id/meta`
2. Get outgoing links: `/api/mind/notes/:id/links`
3. Get incoming links (backlinks): `/api/mind/notes/:id/backlinks`
4. Get note tags: `/api/mind/notes/:id/tags`

### Expected Results
- **HTTP Status**: 200 OK for valid note, 404 if note doesn't exist
- **Response**: List of related resources with appropriate `kind` field

### Verification Steps
```
GET /api/mind/notes/:id/meta

✓ Status is 200
✓ Response kind is "note_meta#list"
✓ Returns array of metadata key-value pairs

GET /api/mind/notes/:id/links

✓ Status is 200
✓ Response kind is "notes_link#list"
✓ Returns array of outgoing links (where note is source)

GET /api/mind/notes/:id/backlinks

✓ Status is 200
✓ Response kind is "notes_link#list"
✓ Returns array of incoming links (where note is destination)

GET /api/mind/notes/:id/tags

✓ Status is 200
✓ Response kind is "tag#list"
✓ Returns array of tags associated with note

GET /api/mind/notes/99999/meta  # Non-existent note

✓ Status is 404
✓ Error message: "Note not found"
```

---

## Workflow 9: Partial Update with Metadata

### User Actions
1. Update a note and include metadata in the request
2. Metadata should be saved separately
3. If metadata save fails, note update should still succeed

### Expected Results
- **Success with metadata**:
  - HTTP Status: 200 OK
  - Note and metadata both updated
  
- **Partial success**:
  - HTTP Status: 207 Multi-Status
  - Note updated but metadata save failed
  - Error details in response

### Verification Steps
```
PUT /api/mind/notes/:id
Headers: If-Match: W/"current_etag"
{
  "title": "Updated Note",
  "meta": {
    "author": "John Doe",
    "rating": 5
  }
}

# Success case
✓ Status is 200
✓ Note title is updated
✓ GET /api/mind/notes/:id/meta returns metadata

# Partial failure case (e.g., invalid meta format)
✓ Status is 207
✓ Note title IS updated
✓ Response includes error details about metadata save failure
✓ Error domain is "note_meta"
```

---

## Workflow 10: Note with Template and Type

### User Actions
1. Create a note with a note_type_id
2. Mark a note as a template with is_template flag
3. Create a note in a specific collection

### Expected Results
- **HTTP Status**: 201 Created
- **Note type**: If provided, associates note with a type
- **Template flag**: If true, note can be used as template for other notes
- **Collection**: Note belongs to specified collection (defaults to root collection id=1)

### Verification Steps
```
POST /api/mind/notes
{
  "title": "Daily Log Template",
  "body": "## Morning\n\n## Afternoon\n\n## Evening",
  "is_template": true,
  "note_type_id": 2,
  "collection_id": 5
}

✓ Status is 201
✓ Note is created with is_template = true
✓ Note has note_type_id = 2
✓ Note belongs to collection_id = 5
✓ GET /api/mind/notes/:id returns these properties
```

---

## Common Error Cases

### Invalid ID Format
```
GET /api/mind/notes/invalid

✓ Status is 400
✓ Error message: "Invalid ID"
```

### Invalid JSON in Request
```
POST /api/mind/notes
{ invalid json

✓ Status is 400
✓ Error message: "Invalid JSON"
```

### Missing Required Fields
```
POST /api/mind/notes
{
  "body": "Content without title"
}

✓ Status is 400
✓ Error message: "Title is required"
```

### Note Not Found
```
GET /api/mind/notes/99999

✓ Status is 404
✓ Error message: "Note not found"
```

---

## Field Selection Behavior

### Supported Fields
All note fields can be individually selected:
- `id`, `uuid`, `title`, `body`, `description`
- `created_at`, `updated_at`
- `note_type_id`, `is_template`, `version`, `etag`

### Field Selection Syntax
```
GET /api/mind/notes/:id?fields=id,title,created_at

✓ Response contains ONLY specified fields
✓ Other fields are omitted (not null, actually omitted)
✓ Works with single note and list endpoints
```

---

## Notes on Response Format

All responses follow consistent structure:

### Success Response
```json
{
  "data": {
    // Entity or list data here
  }
}
```

### Error Response
```json
{
  "error": {
    "code": 400,
    "message": "Human-readable error message",
    "errors": [
      {
        "domain": "note_meta",
        "reason": "saveFailed",
        "message": "Detailed error description"
      }
    ]
  }
}
```

### List Response
```json
{
  "data": {
    "kind": "note#list",
    "items": [],
    "etag": "list_etag_value"
  }
}
```

---

## Summary Checklist for Client Implementation

When implementing a Mindweaver client, ensure you:

- [ ] Handle optimistic locking with If-Match headers
- [ ] Display version numbers for conflict resolution UI
- [ ] Store and reuse ETags for subsequent updates
- [ ] Parse 428 and 409 status codes for update conflicts
- [ ] Handle 207 Multi-Status for partial success scenarios
- [ ] Compose frontmatter into body when displaying notes
- [ ] Extract frontmatter from body when creating/updating notes
- [ ] Support WikiLink syntax `[[Note Title]]` in editor
- [ ] Display backlinks to show note connections
- [ ] Handle field selection for performance optimization
- [ ] Use include parameter to fetch related data efficiently
- [ ] Respect error response format with domain/reason structure
