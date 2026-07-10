<div align="center">
  <h1>jsonserv</h1>
  <p><em>Instant REST API from JSON files</em></p>
  <p>
    <img src="https://img.shields.io/badge/go-1.24-blue" alt="Go 1.24">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT License">
    <img src="https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey" alt="Platform">
  </p>
  <pre>npx jsonserv db.json &nbsp;&nbsp;<span style="color:#555"># Server on http://0.0.0.0:3000</span></pre>
</div>

---

**jsonserv** is a CLI tool that reads a JSON file and generates a full REST API with CRUD
operations for every entity. Like **json-server**, but written in Go — single binary, zero
runtime dependencies, concurrent-safe.

---

## Table of Contents

- [Install](#install)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Query Parameters](#query-parameters)
- [Aliasing Endpoints](#aliasing-endpoints)
- [Image Handling](#image-handling)
- [Team Workflow](#team-workflow)
- [Use Cases](#use-cases)
- [Supported Platforms](#supported-platforms)
- [How It Works](#how-it-works)
- [Full Documentation](#full-documentation)

---

## Install

```bash
npm install -g jsonserv
```

Or run directly without installing:

```bash
npx jsonserv db.json
```

> **No runtime dependencies.** The npm package is a thin JS shim that downloads and
> runs the correct Go binary for your platform.

---

## Quick Start

Create a `db.json` file:

```json
{
  "posts": [
    { "id": "1", "title": "Hello World", "views": 100 },
    { "id": "2", "title": "Second Post", "views": 250 }
  ],
  "comments": [
    { "id": "1", "text": "Nice post!", "postId": "1" },
    { "id": "2", "text": "Thanks for sharing", "postId": "1" }
  ]
}
```

Start the server:

```bash
jsonserv db.json
```

Your API is live:

```bash
curl http://localhost:3000/posts      # list all posts
curl http://localhost:3000/posts/1    # get post by id
```

---

## API Reference

For each entity (top-level array in your JSON), the following endpoints are generated:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/:entity` | List all records |
| `GET` | `/:entity/:id` | Get a single record |
| `POST` | `/:entity` | Create a new record |
| `PUT` | `/:entity/:id` | Replace an entire record |
| `PATCH` | `/:entity/:id` | Partially update a record |
| `DELETE` | `/:entity/:id` | Delete a record |

### Examples

```bash
# Create (auto-generates UUID if no id is provided)
curl -X POST http://localhost:3000/posts \
  -H 'Content-Type: application/json' \
  -d '{"title": "New Post", "views": 50}'

# Create with a custom ID
curl -X POST http://localhost:3000/posts \
  -H 'Content-Type: application/json' \
  -d '{"id": "my-id", "title": "Custom"}'

# Update (full replace)
curl -X PUT http://localhost:3000/posts/1 \
  -H 'Content-Type: application/json' \
  -d '{"title": "Updated", "views": 999}'

# Patch (partial update)
curl -X PATCH http://localhost:3000/posts/1 \
  -H 'Content-Type: application/json' \
  -d '{"views": 150}'

# Delete
curl -X DELETE http://localhost:3000/posts/1
```

---

## Configuration

On first run, `jsonserv` auto-generates a `db.yaml` config file next to your JSON file.
You can customize server settings and entity behaviour.

```yaml
server:
  port: 3000              # HTTP port
  host: "0.0.0.0"         # Bind address
  cors:
    enabled: true          # CORS headers
    origins: ["*"]         # Allowed origins
  logging: true            # Request logging

entities:
  posts:
    filters:
      enabled: true        # ?field=value filtering
    sort:
      enabled: true        # _sort & _order
      default_field: "id"
      default_order: "asc"
    paginate:
      enabled: true        # _page & _limit
      default_page: 1
      default_limit: 10
```

### Generate config without starting the server

```bash
jsonserv init db.json          # generate config from data
jsonserv init db.json --empty  # generate config + empty data template
```

---

## Query Parameters

When listing records, you can filter, search, sort, and paginate.

| Parameter | Example | Description |
|-----------|---------|-------------|
| `?field=value` | `?views=100` | Exact field match |
| `_search` | `?_search=hello` | Case-insensitive substring across string fields |
| `_sort` | `?_sort=views` | Field to sort by |
| `_order` | `?_order=desc` | Sort direction (`asc` or `desc`) |
| `_page` | `?_page=2` | Page number |
| `_limit` | `?_limit=10` | Records per page (response includes `X-Total-Count`) |

```bash
curl "http://localhost:3000/posts?views=100"
curl "http://localhost:3000/posts?_search=hello"
curl "http://localhost:3000/posts?_sort=views&_order=desc"
curl "http://localhost:3000/posts?_page=1&_limit=10"
```

> Each feature can be **enabled or disabled per-entity** in `db.yaml`.

---

## Aliasing Endpoints

Rename an endpoint without changing your JSON structure:

```yaml
# db.yaml
entities:
  posts:
    alias: "articles"   # /posts → /articles
```

Now:
```
GET /articles      → lists posts
GET /posts         → 404
POST /articles     → creates a post
```

---

## Image Handling

jsonserv can automatically compress and resize images when they are sent as base64-encoded
data URIs. Enable it per field in your `db.yaml` config.

### Configuration

```yaml
entities:
  posts:
    imageFields:
      thumbnail:
        maxWidth: 128
        maxHeight: 128
        quality: 30
        format: "jpeg"       # "jpeg" (default) or "png"
      avatar:
        maxWidth: 256
        maxHeight: 256
        quality: 50
        format: "png"
```

| Option | Default | Description |
|--------|---------|-------------|
| `maxWidth` | 128 | Maximum width in pixels (maintains aspect ratio) |
| `maxHeight` | 128 | Maximum height in pixels (maintains aspect ratio) |
| `quality` | 30 | JPEG quality (1-100), ignored for PNG |
| `format` | `jpeg` | Output format: `jpeg` or `png` |

### How It Works

1. Send a record with a field containing a base64 data URI (`data:image/...;base64,...`)
2. jsonserv detects the field is configured in `imageFields`
3. The image is decoded, resized to fit within `maxWidth` x `maxHeight`, and recompressed
4. The compressed data URI replaces the original before saving to `db.json`

Images are processed on **POST**, **PUT**, and **PATCH** requests.

### Example

```bash
# Create a post with a large image — it gets compressed automatically
curl -X POST http://localhost:3000/posts \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Hello",
    "thumbnail": "data:image/jpeg;base64,/9j/4AAQSkZJR..."
  }'
```

The stored record will contain the compressed version of the image.

> **Note:** Only fields configured in `imageFields` are processed. Other fields containing
> base64 data URIs are left untouched. Compression errors are logged but do not block the
> request.

---

## Team Workflow

Share your API structure without sharing data.

```bash
# Alice: generate config + empty data
jsonserv init db.json --empty

# Alice commits db.yaml + db.json (empty arrays)
git add db.yaml db.json
git commit -m "Add API structure"
git push

# Bob: clone and start with fresh data
jsonserv db.yaml     # creates db.json with empty arrays
```

Both get the same endpoints. Bob starts with an empty database.

---

## Use Cases

- **Frontend prototyping** — mock a REST API in seconds without writing a backend
- **Mobile app development** — stand up a fake API while the real one is being built
- **Testing** — spin up isolated test servers with predictable data
- **Demo environments** — share a config file with your team, let everyone run their own instance
- **Quick scripts** — need a temporary CRUD API for a one-off task

---

## Supported Platforms

| OS | x64 | ARM64 |
|----|:---:|:-----:|
| Linux | ✅ | ✅ |
| macOS | ✅ | ✅ (Apple Silicon) |
| Windows | ✅ | ✅ |

---

## How It Works

1. You run `jsonserv db.json`
2. The JS shim detects your OS and architecture
3. It spawns the correct prebuilt Go binary
4. The binary reads `db.json`, registers routes, and starts an HTTP server
5. All mutations are persisted to `db.json` immediately
6. If no `db.yaml` exists, one is auto-generated with inferred schemas

The Go binary is compiled with `CGO_ENABLED=0` and `-ldflags="-s -w"` for a small,
statically linked binary (~6MB per platform).

---

## Full Documentation

See **[docs/GUIDE.md](docs/GUIDE.md)** for the complete guide covering:

- [All CLI commands in detail](docs/GUIDE.md#3-cli-commands)
- [Complete API endpoint documentation](docs/GUIDE.md#5-api-endpoints)
- [Advanced YAML configuration](docs/GUIDE.md#7-the-yaml-config-file)
- [End-to-end walkthrough](docs/GUIDE.md#9-end-to-end-example)
