# Mind V3 API - Bruno Collection

This Bruno collection provides API testing for the Mind V3 service (PKM/Notes) with Connect-RPC.

## Setup

### Prerequisites
- **Bruno Desktop App**: Download from [usebruno.com](https://www.usebruno.com/)
- **Bruno CLI**: Install via `npm install -g @usebruno/cli`
- **Mind Service Running**: Start the service on `http://localhost:9421`

### Opening the Collection

#### In Bruno GUI:
1. Open Bruno Desktop App
2. Click "Open Collection"
3. Navigate to `tests/api/mind-v3/`
4. Select the folder

#### Using Bruno CLI:
```bash
# From project root
cd tests/api/mind-v3
```

## Collection Structure

```
mind-v3/
├── bruno.json                      # Collection config
├── environments/
│   └── local.bru                   # Local environment (localhost:9421)
├── templates/                      # Template API (5 requests)
│   ├── 1-create-template.bru       # Creates template + saves ID
│   ├── 2-get-template.bru          # Gets template by ID
│   ├── 3-list-templates.bru        # Lists all templates
│   ├── 4-update-template.bru       # Updates template
│   └── 5-delete-template.bru       # Deletes template
└── collections/                    # Collection API (7 requests)
    ├── 1-create-collection.bru     # Creates collection + saves ID
    ├── 2-get-collection.bru        # Gets collection by ID
    ├── 3-list-collections.bru      # Lists all collections
    ├── 4-update-collection.bru     # Updates collection
    ├── 5-list-children.bru         # Lists child collections
    ├── 6-get-tree.bru              # Gets collection tree
    └── 7-delete-collection.bru     # Deletes collection
```

## Usage

### GUI Testing (Manual)

1. **Select Environment**: Choose `local` from the environment dropdown
2. **Run Single Request**: Click any `.bru` file and click "Send"
3. **Run Folder**: Right-click folder (e.g., `templates`) → "Run Folder"
4. **View Response**: See JSON response in right panel
5. **Check Tests**: View test results at bottom

### CLI Testing (Automated)

#### Run Single Request
```bash
bru run templates/1-create-template.bru --env local
```

#### Run Resource Folder
```bash
# Test all template endpoints
bru run templates --env local

# Test all collection endpoints
bru run collections --env local
```

#### Run Entire Collection
```bash
# From tests/api/mind-v3/
bru run . --env local

# Or from project root
bru run tests/api/mind-v3 --env local
```

#### With Reporters
```bash
# JSON output
bru run templates --env local --reporter json

# JUnit XML (for CI)
bru run templates --env local --reporter junit --output results.xml

# HTML report
bru run templates --env local --reporter html --output report.html
```

## Request Chaining

Requests use **environment variables** to chain together:

1. **Create requests** save IDs to environment:
   ```javascript
   // In 1-create-template.bru
   script:post-response {
     const templateId = res.body.name.split("/")[1];
     bru.setEnvVar("templateId", templateId);
   }
   ```

2. **Subsequent requests** use saved IDs:
   ```json
   // In 2-get-template.bru
   {
     "name": "templates/{{templateId}}"
   }
   ```

This allows you to run entire resource folders sequentially:
```bash
# Creates → Gets → Lists → Updates → Deletes
bru run templates --env local
```

## Test Assertions

Each request includes test assertions:

```javascript
tests {
  test("Status is 200", function() {
    expect(res.status).to.equal(200);
  });
  
  test("Response has expected fields", function() {
    expect(res.body).to.have.property("name");
  });
}
```

View test results in GUI or CLI output.

## Dynamic Data

Requests use Bruno's built-in variables for unique data:

- `{{$timestamp}}` - Unix timestamp
- `{{$guid}}` - UUID
- `{{$randomInt}}` - Random integer
- `{{$randomProductName}}` - Random product name

Example:
```json
{
  "template": {
    "name": "Daily Note - {{$timestamp}}",
    "content": "# Note content"
  }
}
```

## Environment Variables

### Defined in `environments/local.bru`:
- `baseUrl`: `http://localhost:9421`
- `templateId`: (set by create-template request)
- `collectionId`: (set by create-collection request)

### Add More Variables:
Edit `environments/local.bru`:
```bru
vars {
  baseUrl: http://localhost:9421
  templateId: 
  collectionId:
  noteId:        # Add for notes testing
  tagId:         # Add for tags testing
}
```

## Troubleshooting

### Server Not Running
```bash
# Error: connect ECONNREFUSED 127.0.0.1:9421
# Solution: Start the Mind service
./mindweaver
```

### Environment Variable Not Set
```bash
# Error: templateId is undefined
# Solution: Run create request first
bru run templates/1-create-template.bru --env local
bru run templates/2-get-template.bru --env local
```

### Request Fails
1. Check response tab in GUI
2. View detailed error in CLI: `bru run <file> --env local --verbose`
3. Verify JSON syntax in body

## Next Steps

### Add More Resources
When ready to test other V3 endpoints:

1. Create folder: `mkdir tests/api/mind-v3/notes`
2. Add requests: `notes/1-create-note.bru`, etc.
3. Update environment: Add `noteId` variable
4. Run: `bru run notes --env local`

### Resources to Add:
- [ ] Tags (5 CRUD requests)
- [ ] NoteTypes (5 CRUD requests)  
- [ ] Notes (5 requests + metadata)
- [ ] NoteMeta (1 request - read-only sub-resource)
- [ ] Search (1 request)

## Contributing

When adding new requests:
1. Use numeric prefixes for execution order: `1-create-*.bru`, `2-get-*.bru`, etc.
2. Save resource IDs in post-response scripts
3. Add test assertions
4. Use dynamic variables for unique data
5. Follow existing naming conventions

## Resources

- [Bruno Docs](https://docs.usebruno.com/)
- [Bruno CLI](https://docs.usebruno.com/cli/overview)
- [Connect-RPC Protocol](https://connectrpc.com/)
- Mind V3 API: `proto/mind/v3/*.proto`
