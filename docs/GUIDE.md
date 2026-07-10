# jsonserv — REST API Server from JSON

`jsonserv` is a lightweight CLI tool that reads a JSON data file and instantly generates a
full REST API with CRUD operations for every entity in it. It is written in Go — a single
binary, no runtime dependencies.

---

## Table of Contents

1. [Installation](#1-installation)
2. [Quick Start](#2-quick-start)
3. [CLI Commands](#3-cli-commands)
   - [jsonserv \<file.json\>](#31-jsonserv-filejson)
   - [jsonserv \<file.yaml\>](#32-jsonserv-fileyaml)
   - [jsonserv init \<file.json\>](#33-jsonserv-init-filejson)
   - [jsonserv init \<file.json\> --empty](#34-jsonserv-init-filejson---empty)
4. [The JSON Data File](#4-the-json-data-file)
5. [API Endpoints](#5-api-endpoints)
   - [List all records](#51-list-all-records-get-entity)
   - [Get one record](#52-get-one-record-get-entityid)
   - [Create a record](#53-create-a-record-post-entity)
   - [Update a record (full replace)](#54-update-a-record-put-entityid)
   - [Patch a record (partial update)](#55-patch-a-record-patch-entityid)
   - [Delete a record](#56-delete-a-record-delete-entityid)
6. [Query Parameters (Filtering &amp; Search)](#6-query-parameters)
   - [Field filtering](#61-field-filtering)
   - [Search](#62-search-_search)
   - [Sorting](#63-sorting-_sort--_order)
   - [Pagination](#64-pagination-_page--_limit)
7. [The YAML Config File](#7-the-yaml-config-file)
   - [Server settings](#71-server-settings)
   - [Entity configuration](#72-entity-configuration)
   - [Aliasing endpoints](#73-aliasing-endpoints)
   - [Schema](#74-schema)
   - [Image Handling](#75-image-handling)
8. [Team Workflow — Sharing Config](#8-team-workflow)
9. [End-to-End Example](#9-end-to-end-example)

---

## 1. Installation

```bash
go build -o jsonserv ./cmd/myserv
```

This produces a single static binary. Place it anywhere in your `$PATH`:

```bash
sudo mv jsonserv /usr/local/bin/
```

No other dependencies are needed at runtime.

---

## 2. Quick Start

Create a `db.json` file:

```json
{
  "posts": [
    { "id": "1", "title": "Hello World", "views": 100 },
    { "id": "2", "title": "Second Post", "views": 250 }
  ],
  "comments": [
    { "id": "1", "text": "Nice post!", "postId": "1" }
  ]
}
```

Start the server:

```bash
jsonserv db.json
```

Output:

```
config auto-generated: db.yaml
  route: posts -> /posts
  route: comments -> /comments
jsonserv running on http://0.0.0.0:3000
```

Your API is live. The first time you run `jsonserv`, it automatically generates a
`db.yaml` config file alongside your JSON.

---

## 3. CLI Commands

### 3.1 `jsonserv <file.json>`

Start the server using a JSON data file.

- Loads every top-level array in the JSON as an entity.
- If a `.yaml` file with the same base name exists (e.g. `db.yaml`), it loads it
  for configuration. Otherwise, it **auto-generates** one with default settings
  and inferred schemas.

```bash
jsonserv db.json
```

### 3.2 `jsonserv <file.yaml>`

Start the server from a YAML config file only.

- Reads the entity schemas from the YAML.
- If no corresponding `.json` file exists, it **creates one** with empty arrays
  for every entity and starts the server.
- If a `.json` file already exists, it loads it normally.

```bash
jsonserv db.yaml
```

### 3.3 `jsonserv init <file.json>`

Generate a YAML config file from an existing JSON data file, then exit.

- Analyzes the data to infer field names and types.
- Writes a `db.yaml` with default settings for every entity.
- Does **not** start the server.

```bash
jsonserv init db.json
```

### 3.4 `jsonserv init <file.json> --empty`

Same as `init`, but also empties the JSON data file (leaves only empty arrays).

- Useful when you want to share the structure with a team without sharing your data.

```bash
jsonserv init db.json --empty
```

---

## 4. The JSON Data File

The JSON file serves as both the data source and the persistent storage.

- Every **top-level key** whose value is an **array of objects** becomes an entity.
- Each object in the array is a record.
- Objects can have any fields — no schema is enforced at the API level.
- Every record should have an `"id"` field. If missing on creation, it is auto-generated.

```json
{
  "posts": [
    { "id": "1", "title": "Hello", "views": 100 },
    { "id": "2", "title": "World", "views": 200 }
  ],
  "users": [
    { "id": "abc", "name": "Alice", "role": "admin" }
  ]
}
```

After every mutating request (POST, PUT, PATCH, DELETE), the file is **written back**
to disk immediately, exactly like `json-server`.

---

## 5. API Endpoints

For each entity, `jsonserv` generates six endpoints.

### 5.1 List all records: `GET /:entity`

```bash
curl http://localhost:3000/posts
```

Returns a JSON array of all records.

### 5.2 Get one record: `GET /:entity/:id`

```bash
curl http://localhost:3000/posts/1
```

Returns a single JSON object, or `404` with `{"error":"not found"}`.

### 5.3 Create a record: `POST /:entity`

```bash
curl -X POST http://localhost:3000/posts \
  -H 'Content-Type: application/json' \
  -d '{"title": "New Post", "views": 50}'
```

- If the body contains an `"id"` field, it is used as-is.
- If no `"id"` is provided, a **UUID v4** is auto-generated.
- Returns `201` with the created record.

#### Custom ID example

```bash
curl -X POST http://localhost:3000/posts \
  -H 'Content-Type: application/json' \
  -d '{"id": "my-custom-id", "title": "Custom"}'
```

### 5.4 Update a record (full replace): `PUT /:entity/:id`

```bash
curl -X PUT http://localhost:3000/posts/1 \
  -H 'Content-Type: application/json' \
  -d '{"title": "Replaced", "views": 999}'
```

- Replaces the entire record. The `id` from the URL is preserved regardless of the body.
- Returns `200` with the updated record, or `404` if the record does not exist.

### 5.5 Patch a record (partial update): `PATCH /:entity/:id`

```bash
curl -X PATCH http://localhost:3000/posts/1 \
  -H 'Content-Type: application/json' \
  -d '{"views": 150}'
```

- Merges the provided fields into the existing record.
- The `"id"` field cannot be changed via PATCH.
- Returns `200` with the updated record, or `404` if not found.

### 5.6 Delete a record: `DELETE /:entity/:id`

```bash
curl -X DELETE http://localhost:3000/posts/1
```

- Returns `204 No Content` on success, or `404` if not found.

---

## 6. Query Parameters

When listing records (`GET /:entity`), you can use query parameters to filter, search,
sort, and paginate. These features are **enabled or disabled per entity** in the YAML
config file.

### 6.1 Field Filtering

Filter by exact field value:

```bash
curl "http://localhost:3000/posts?views=100"
```

```bash
curl "http://localhost:3000/posts?title=Hello%20World"
```

Multiple filters can be combined:

```bash
curl "http://localhost:3000/posts?title=Hello&views=100"
```

### 6.2 Search: `_search`

Search across all string fields (case-insensitive substring match):

```bash
curl "http://localhost:3000/posts?_search=hello"
```

Returns records where any string field contains "hello".

### 6.3 Sorting: `_sort` & `_order`

```bash
curl "http://localhost:3000/posts?_sort=views&_order=desc"
```

- `_sort`: the field to sort by.
- `_order`: `asc` (default) or `desc`.
- Works on both numeric and string fields.

### 6.4 Pagination: `_page` & `_limit`

```bash
curl "http://localhost:3000/posts?_page=1&_limit=10"
```

- `_page`: page number (starts at 1).
- `_limit`: records per page.
- The response includes the header `X-Total-Count` with the total number of records.

---

## 7. The YAML Config File

The YAML file (`db.yaml` by default) controls server behavior and entity-level settings.
It is **auto-generated** on first run but can be edited manually.

### 7.1 Server Settings

```yaml
server:
  port: 3000              # HTTP port (default: 3000)
  host: "0.0.0.0"         # Bind address (default: 0.0.0.0)
  cors:
    enabled: true          # Enable CORS headers
    origins: ["*"]         # Allowed origins
  logging: true            # Request logging to stdout
```

### 7.2 Entity Configuration

Each entity can have its own settings:

```yaml
entities:
  posts:
    filters:
      enabled: true        # Enable ?field=value filtering
    sort:
      enabled: true        # Enable _sort & _order
      default_field: "id"  # Default sort field
      default_order: "asc" # Default sort direction
    paginate:
      enabled: true        # Enable _page & _limit
      default_page: 1      # Default page
      default_limit: 10    # Default records per page
```

### 7.3 Aliasing Endpoints

Rename an endpoint without changing your JSON data:

```yaml
entities:
  posts:
    alias: "articles"      # Accessible at /articles instead of /posts
```

Now:

| Endpoint | Result |
|----------|--------|
| `GET /articles` | Lists posts |
| `GET /posts` | 404 Not Found |

All CRUD operations, filtering, sorting, and pagination work through the aliased path.

### 7.4 Schema

The schema section is **auto-inferred** from your data. It documents the fields and
their types. It is used when generating empty data from a YAML-only start.

```yaml
entities:
  posts:
    schema:
      id: "string"
      title: "string"
      views: "number"
```

Supported types: `string`, `number`, `boolean`.

### 7.5 Image Handling

jsonserv can automatically **compress and resize** images sent as base64 data URIs.
When enabled, images are processed transparently on POST, PUT, and PATCH requests.

#### Enabling Image Fields

Add an `imageFields` map to any entity in your YAML config. Each key is a field name,
and the value specifies compression settings:

```yaml
entities:
  users:
    imageFields:
      avatar:
        maxWidth: 256
        maxHeight: 256
        quality: 80
        format: "jpeg"
      thumbnail:
        maxWidth: 64
        maxHeight: 64
        quality: 30
        format: "png"
```

#### Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maxWidth` | int | 128 | Maximum width in pixels (aspect ratio is preserved) |
| `maxHeight` | int | 128 | Maximum height in pixels (aspect ratio is preserved) |
| `quality` | int | 30 | JPEG quality 1-100 (ignored for PNG) |
| `format` | string | `"jpeg"` | Output format: `"jpeg"` or `"png"` |

#### How It Works

1. A request arrives with a JSON body containing a field like `"avatar": "data:image/jpeg;base64,..."`
2. jsonserv checks if that field is listed in `imageFields` for the entity
3. The image is **decoded**, **resized** to fit within `maxWidth × maxHeight`, and **recompressed**
4. The compressed base64 data URI replaces the original in the body
5. The modified record is stored in `db.json`

> Images that are already smaller than `maxWidth × maxHeight` are only recompressed,
> not upscaled.

#### Supported Formats

- **JPEG** — output uses the `quality` setting
- **PNG** — lossless encoding (quality is ignored)
- Input can be any format supported by Go's `image` package (JPEG, PNG, GIF, BMP)

#### Sending Images

Include the image as a base64 data URI in your JSON body:

```bash
curl -X POST http://localhost:3000/users \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Mr. Meow",
    "title": "Chief Napping Officer",
    "bio": "Professional window sitter, amateur mouse chaser.",
    "avatar": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
  }'
```

The response contains the **compressed** version:

```json
{
  "id": "mr-meow-001",
  "name": "Mr. Meow",
  "avatar": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
}
```

#### Error Handling

If compression fails (e.g. corrupted image data), the error is logged but the request
**still succeeds** — the original value is stored unchanged.

---

## 8. Team Workflow

Share only the API structure with your team, without sharing your data.

### Step 1: Alice generates the config

```bash
# Alice has data in db.json
jsonserv init db.json --empty
```

This produces:
- `db.yaml` — full config with entity schemas and default settings
- `db.json` — emptied (empty arrays, structure preserved)

### Step 2: Alice commits & pushes

```bash
git add db.yaml db.json
git commit -m "Add API config for posts and users"
git push
```

### Step 3: Bob starts the server

```bash
# Bob clones the repo
jsonserv db.yaml
```

Output:

```
data file created: db.json
  route: posts -> /posts
  route: users -> /users
jsonserv running on http://0.0.0.0:3000
```

Bob now has:
- The same API endpoints as Alice
- An empty `db.json` ready to receive data
- The ability to customize the YAML without affecting anyone else

---

## 9. End-to-End Example

```bash
# 1. Create a data file
cat > db.json << 'EOF'
{
  "products": [
    { "id": "1", "name": "Widget", "price": 9.99 },
    { "id": "2", "name": "Gadget", "price": 24.99 },
    { "id": "3", "name": "Doohickey", "price": 4.99 }
  ]
}
EOF

# 2. Start the server
jsonserv db.json &
# Output: config auto-generated, server running on :3000

# 3. List all products
curl http://localhost:3000/products

# 4. Get one product
curl http://localhost:3000/products/1

# 5. Create a product (auto-generated UUID)
curl -X POST http://localhost:3000/products \
  -H 'Content-Type: application/json' \
  -d '{"name": "New Item", "price": 14.99}'

# 6. Filter products by price
curl "http://localhost:3000/products?price=9.99"

# 7. Sort products by price descending
curl "http://localhost:3000/products?_sort=price&_order=desc"

# 8. Search products containing "gad"
curl "http://localhost:3000/products?_search=gad"

# 9. Paginate (page 1, 2 per page)
curl "http://localhost:3000/products?_page=1&_limit=2"

# 10. Generate config for sharing
jsonserv init db.json --empty

# 11. Stop the server
kill %1
```
