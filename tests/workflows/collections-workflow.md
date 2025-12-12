# Collections Workflow

## Overview

Collections organize notes hierarchically. Each note belongs to exactly one collection.

---

## Flow 1: Create Collections

**What happens**: User creates nested collections to organize notes by topic/project.

**Test**:
```bash
# Create top-level "Books" collection
curl -X POST http://localhost:9421/api/mind/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "Books", "description": "Book notes and reviews"}'

# Create "Work" collection
curl -X POST http://localhost:9421/api/mind/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "Work", "description": "Work-related notes"}'

# Create nested "Projects" under "Work" (parent_id = 2)
curl -X POST http://localhost:9421/api/mind/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "Projects", "parent_id": 2, "description": "Active projects"}'
```

**Expected**:
- Collections created with auto-generated paths
- "Books" path: `/Books`
- "Work" path: `/Work`
- "Projects" path: `/Work/Projects`

---

## Flow 2: List Collections

**What happens**: User views all collections or filters by parent.

**Test**:
```bash
# List all collections
curl http://localhost:9421/api/mind/collections

# List only top-level collections (no parent)
curl 'http://localhost:9421/api/mind/collections?parent_id=null'

# List children of "Work" collection (parent_id = 2)
curl http://localhost:9421/api/mind/collections?parent_id=2
```

**Expected**:
- Returns collections with id, name, path, parent_id
- Filtering works correctly

---

## Flow 3: Create Note in Collection

**What happens**: User creates a note in a specific collection.

**Test**:
```bash
# Create note in "Books" collection (id = 1)
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Clean Code\", \"collection_id\": 1, \"body\": $(cat testdata/markdown/book-clean-code.md | jq -Rs .)}"

# Create note in "Projects" collection (id = 3)
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Project Alpha\", \"collection_id\": 3, \"body\": $(cat testdata/markdown/project-alpha.md | jq -Rs .)}"
```

**Expected**:
- Notes created in specified collections
- Default collection_id = 1 (root) if not specified

---

## Flow 4: Move Note to Different Collection

**What happens**: User moves a note from one collection to another.

**Test**:
```bash
# Get note etag first
ETAG=$(curl -s http://localhost:9421/api/mind/notes/1 | jq -r '.data.etag')

# Move "Clean Code" from "Books" (1) to "Work" (2)
curl -X PUT http://localhost:9421/api/mind/notes/1 \
  -H "Content-Type: application/json" \
  -H "If-Match: $ETAG" \
  -d '{"collection_id": 2}'
```

**Expected**:
- Note moved to new collection
- WikiLinks remain intact (links are by title, not collection-specific)
- Version incremented

---

## Flow 5: Duplicate Title in Different Collections

**What happens**: Same title allowed in different collections.

**Test**:
```bash
# Create "Meeting Notes" in "Books" collection (id = 1)
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d '{"title": "Meeting Notes", "collection_id": 1, "body": "Book club meeting"}'

# Create another "Meeting Notes" in "Work" collection (id = 2) - should succeed
curl -X POST http://localhost:9421/api/mind/notes \
  -H "Content-Type: application/json" \
  -d '{"title": "Meeting Notes", "collection_id": 2, "body": "Sprint planning"}'
```

**Expected**:
- Both notes created successfully
- Title uniqueness is per-collection, not global
- Constraint: `UNIQUE (collection_id, title)`

---

## Flow 6: Update Collection

**What happens**: User renames or moves a collection.

**Test**:
```bash
# Rename "Books" to "Reading"
curl -X PUT http://localhost:9421/api/mind/collections/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Reading", "description": "Books and articles"}'
```

**Expected**:
- Collection renamed
- Path updated automatically
- Notes in collection unaffected

---

## Flow 7: Delete Empty Collection

**What happens**: User deletes a collection that has no notes.

**Test**:
```bash
# Create empty collection
curl -X POST http://localhost:9421/api/mind/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "Temp"}'

# Delete it
curl -X DELETE http://localhost:9421/api/mind/collections/4
```

**Expected**:
- Collection deleted successfully
- Child collections deleted (cascade)

---

## Flow 8: Delete Collection with Notes

**What happens**: When deleting a collection with notes, notes move to default collection (id=1).

**Test**:
```bash
# Delete "Books" collection which has notes
curl -X DELETE http://localhost:9421/api/mind/collections/1
```

**Expected**:
- Collection deleted
- Notes previously in "Books" now have collection_id = 1 (root/default)
- Constraint: `ON DELETE SET DEFAULT`

---

## Flow 9: Get Collection Stats

**What happens**: User checks how many notes are in a collection.

**Test**:
```bash
curl http://localhost:9421/api/mind/collections/1/stats
```

**Expected**:
- Returns note count for that collection
- Useful for UI to show collection sizes

---

## Issues Found

### Run 1 - 2025-12-09

#### Collections Working Correctly
- ✅ Collections created successfully with auto-generated paths
- ✅ Nested collections work (Projects under Books: `/books/projects`)
- ✅ List all collections works
- ✅ Creating notes in specific collections works
- ✅ Same title in different collections allowed (correct)
- ✅ Moving note to collection with duplicate title blocked (correct)

#### Related Issues from Notes Workflow
See notes-workflow.md Issue #5 - Moving note with duplicate title needs better error message

### Summary
Collections functionality working as expected. Only improvement needed is error messaging (tracked in notes workflow).
