# Notes Workflow

## Test Data Setup

All test data is in `/testdata/markdown/`:
- `book-clean-code.md` - Book note with author, ISBN, rating, WikiLinks
- `daily-log-2024-03-15.md` - Daily note with mood, tasks, WikiLinks
- `john-doe.md` - Person note with role, email
- `project-alpha.md` - Project note with status, priority, team members
- `meeting-notes-2024-03.md` - Meeting note with attendees, action items
- `quick-thoughts.md` - Simple note with WikiLinks, no frontmatter

---

## Flow 1: Create Notes from Files

**What happens**: User creates notes from markdown files. System extracts frontmatter and WikiLinks.

**Test**:
```bash
# Create book note
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Clean Code\", \"body\": $(cat testdata/markdown/book-clean-code.md | jq -Rs .)}"

# Create person note  
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"John Doe\", \"body\": $(cat testdata/markdown/john-doe.md | jq -Rs .)}"

# Create project note
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Project Alpha\", \"body\": $(cat testdata/markdown/project-alpha.md | jq -Rs .)}"
```

**Expected**:
- Notes created successfully
- Frontmatter extracted from YAML blocks
- WikiLinks parsed (e.g., `[[api-design]]`, `[[john-doe]]`)
- Pending links created for non-existent notes (dest_id = null)

**Verify**: Check links were extracted
```bash
curl http://localhost:9421/api/mind/notes/1/links
```

---

## Flow 2: WikiLink Resolution

**What happens**: When a note is created, any pending links to it (by title) should resolve automatically.

**Test**:
```bash
# book-clean-code.md links to [[api-design]] - currently pending
# Now create the target note
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"API Design\", \"body\": $(cat testdata/markdown/api-design.md | jq -Rs .)}"
```

**Expected**:
- Link from "Clean Code" to "API Design" now has dest_id populated
- Backlinks work: GET /notes/{api-design-id}/backlinks shows "Clean Code"

**Verify**:
```bash
curl http://localhost:9421/api/mind/notes/1/links  # Should show resolved dest_id
curl http://localhost:9421/api/mind/notes/4/backlinks  # Should show backlink from note 1
```

---

## Flow 3: Duplicate Detection

**What happens**: User tries to create a note with a title that already exists in the same collection. System rejects it.

**Test**:
```bash
# Try to create duplicate "Clean Code" in same collection
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d '{"title": "Clean Code", "body": "Different content"}'
```

**Expected**:
- Request fails with error
- Error indicates duplicate title in collection
- Existing note is unchanged

---

## Flow 4: List Notes by Metadata

**What happens**: User filters notes by metadata key (from frontmatter or enhanced metadata).

**Test**:
```bash
# Find all notes with "author" field (books)
curl http://localhost:9421/api/mind/notes?meta_key=author

# Find all notes with "status" field (projects)
curl http://localhost:9421/api/mind/notes?meta_key=status

# Find all notes with "type" field
curl http://localhost:9421/api/mind/notes?meta_key=type
```

**Expected**:
- Only notes containing that metadata key are returned
- Works for frontmatter metadata
- Works for enhanced metadata added via API

---

## Flow 5: Update Note Body

**What happens**: User updates note content. WikiLinks are re-parsed. Old links removed, new links added.

**Test**:
```bash
# Get current note and etag
ETAG=$(curl -s http://localhost:9421/api/mind/notes/1 | jq -r '.data.etag')

# Update with new content
curl -X PUT http://localhost:9421/api/mind/notes/1 \
  -H "Content-Type: application/json" \
  -H "If-Match: $ETAG" \
  -d '{"body": "# Updated\n\nNow links to [[New Note]] and [[Another Note]]."}'
```

**Expected**:
- Old links deleted
- New links from updated body created
- Version incremented
- New etag returned

**Verify**:
```bash
curl http://localhost:9421/api/mind/notes/1/links  # Should show only new links
```

---

## Flow 6: Get Note with Relations

**What happens**: User requests a note with all related data in one call.

**Test**:
```bash
curl 'http://localhost:9421/api/mind/notes/1?include=meta,tags,links'
```

**Expected**:
- Response includes `meta` array with key-value pairs
- Response includes `tags` array
- Response includes `links` array (outgoing links)
- Single request, no N+1 queries

---

## Flow 7: Delete Note

**What happens**: User deletes a note. All associated data is cleaned up.

**Test**:
```bash
curl -X DELETE http://localhost:9421/api/mind/notes/1
```

**Expected**:
- Note deleted
- Metadata cascade deleted
- Outgoing links cascade deleted
- Incoming links updated (dest_id set to null or deleted)
- Tags unlinked

**Verify**:
```bash
curl http://localhost:9421/api/mind/notes/1  # Should return 404
```

---

## Issues Found

### Run 1 - 2025-12-09

#### Issue #1: WikiLink Resolution NOT Working
- **Flow**: Flow 2 - WikiLink Resolution
- **Description**: When a target note is created, existing pending links to it do NOT resolve automatically
- **Expected**: Link from "Clean Code" (id=1) to "api-design" should have `dest_id: 4` and `resolved: 1` after creating "API Design" note
- **Actual**: Link still shows `dest_id: null`, `resolved: 0`, and backlinks endpoint returns empty array
- **Impact**: HIGH - Core feature broken, backlinks won't work
- **Test**:
  ```bash
  # Created "Clean Code" with link to [[api-design]] - pending
  # Created "API Design" note
  # Link from Clean Code still unresolved
  curl http://localhost:9421/api/mind/notes/1/links | jq '.data.items[] | select(.dest_title == "api-design")'
  # Shows: resolved: 0, no dest_id
  ```

#### Issue #2: Duplicate Detection Error Message
- **Flow**: Flow 3 - Duplicate Detection
- **Description**: Duplicate title detection works but returns generic error message
- **Expected**: Error should indicate "Note with title 'Clean Code' already exists in this collection"
- **Actual**: Generic "Failed to create note" message
- **Impact**: MEDIUM - Works but not user-friendly
- **Test**:
  ```bash
  curl -X POST http://localhost:9421/api/mind/notes -d '{"title": "Clean Code", "body": "x"}'
  # Returns: {"error": {"code": 500, "message": "Failed to create note"}}
  ```

#### Issue #3: WikiLinks NOT Parsed on Update
- **Flow**: Flow 5 - Update Note Body
- **Description**: When updating note body, old links are deleted but new WikiLinks are NOT parsed/stored
- **Expected**: After updating body to "[[New Note]] and [[Another Note]]", should have 2 new links
- **Actual**: All links deleted, 0 new links created
- **Impact**: HIGH - Can't update note content with new links
- **Test**:
  ```bash
  # Updated note 1 body with [[New Note]] and [[Another Note]]
  curl http://localhost:9421/api/mind/notes/1/links
  # Shows: 0 links (should have 2)
  ```

#### Issue #4: Include Parameter for Links Not Working
- **Flow**: Flow 6 - Get Note with Relations
- **Description**: When using `?include=links`, the links are not populated in response
- **Expected**: Response should include `links` array with link data
- **Actual**: Direct endpoint `/notes/2/links` works, but `?include=links` doesn't populate
- **Impact**: MEDIUM - Can workaround with direct endpoint but API inconsistent
- **Test**:
  ```bash
  curl 'http://localhost:9421/api/mind/notes/2?include=links' | jq '.data.links'
  # Returns: null (but /notes/2/links has 6 links)
  ```

#### Issue #5: Moving Note to Collection with Duplicate Title
- **Flow**: Collections - Move Note
- **Description**: Moving note to collection that has same title fails with generic error
- **Expected**: Clear error "Cannot move: note with title 'System Architecture' already exists in target collection"
- **Actual**: Generic "Failed to update note"
- **Impact**: MEDIUM - Works correctly but error message not helpful
- **Test**:
  ```bash
  # Created "System Architecture" in both collection 3 and 4
  # Try to move note from collection 3 to 4
  curl -X PUT http://localhost:9421/api/mind/notes/5 -d '{"collection_id": 4}'
  # Returns: {"error": {"code": 500, "message": "Failed to update note"}}
  ```

### Summary
- **Critical Issues**: 2 (WikiLink resolution, WikiLink parsing on update)
- **Medium Issues**: 3 (Error messages not descriptive)
- **Total Issues**: 5
