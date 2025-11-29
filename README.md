# GraphQL Personal Knowledge Management (PKM) System

A GraphQL-based Personal Knowledge Management system with AI-powered search and smart linking, built using Go and Groq.

## Features

- **Note Management**: Create, read, update, and delete notes with rich metadata
- **Smart Linking**: Bidirectional links between notes with backlink tracking
- **AI-Powered Search**: Semantic, hybrid, and smart search capabilities
- **Tagging System**: Organize notes with flexible tagging
- **Real-time GraphQL API**: Efficient queries and mutations
- **AI Insights**: Automated connection suggestions and gap detection
- **Persistent Storage**: MySQL database with embedding caching
- **Cost Optimization**: Intelligent embedding cache to reduce API costs

## Tech Stack

- **Backend**: Go 1.21+
- **GraphQL**: gqlgen
- **AI**: Groq API integration (ultra-fast inference)
- **Database**: MySQL 5.7+ / MariaDB 10.2+
- **Caching**: MySQL-based embedding cache
- **Configuration**: godotenv

## Installation

### Requirements

- Go 1.21 or newer
- MySQL 5.7+ or MariaDB 10.2+
- Groq API key (optional, for AI features)

### Steps

1. **Clone repository**:
```bash
git clone <repository>
cd graphql-pkm
```

2. **Install dependencies**:
```bash
go mod tidy
```

3. **Set up MySQL database**:
```bash
mysql -u root -p
CREATE DATABASE pkm;
```

4. **Configure environment**:
```bash
cp .env.example .env
```

Edit `.env` with your settings:
```bash
PORT=8080
ENVIRONMENT=development
apiKey=gsk_your_groq_api_key_here
apiUrl=https://api.groq.com/openai/v1
defaultModel=llama-3.3-70b-versatile
dbUrl=root:password@tcp(localhost:3306)/pkm?parseTime=true
```

5. **Generate GraphQL code**:
```bash
go run github.com/99designs/gqlgen generate
```

6. **Run server**:
```bash
go run cmd/server/main.go
```

You should see:
```
✓ MySQL database connected successfully
✓ Database tables verified/created successfully
🚀 GraphQL server running on http://localhost:8080
🎮 Playground at http://localhost:8080
```

7. **Open GraphQL Playground**:  
   http://localhost:8080

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENVIRONMENT` | Environment mode | `development` |
| `apiKey` | Groq API key | `""` (AI disabled) |
| `apiUrl` | Groq API endpoint | `https://api.groq.com/openai/v1` |
| `defaultModel` | AI model to use | `llama-3.3-70b-versatile` |
| `dbUrl` | MySQL connection string | `root:password@tcp(localhost:3306)/pkm?parseTime=true` |

### Supported AI Models (Groq)

Groq provides ultra-fast inference speeds with these models:

- `llama-3.3-70b-versatile` (Recommended - Fast & capable, 128K context)
- `llama-3.1-70b-versatile` (Excellent for most tasks, 128K context)
- `llama-3.1-8b-instant` (Fastest response time, 128K context)
- `mixtral-8x7b-32768` (Good for reasoning tasks, 32K context)
- `gemma2-9b-it` (Efficient and fast, 8K context)

## API Usage

### Note Management

#### Create Note
```graphql
mutation {
  createNote(input: {
    title: "GraphQL Basics"
    content: "GraphQL is a query language for APIs that provides a complete and understandable description of the data."
    tags: ["graphql", "api", "learning"]
  }) {
    id
    title
    content
    tags
    createdAt
    updatedAt
  }
}
```

#### Get All Notes
```graphql
query {
  notes {
    id
    title
    content
    tags
    createdAt
    updatedAt
  }
}
```

#### Get Single Note
```graphql
query {
  note(id: "your_note_id") {
    id
    title
    content
    tags
    summary
    keyConcepts
    links {
      id
      targetNoteId
      description
    }
    backlinks {
      id
      sourceNoteId
      description
    }
  }
}
```

#### Update Note
```graphql
mutation {
  updateNote(id: "note_id", input: {
    title: "Updated GraphQL Basics"
    content: "Updated content with more details"
    tags: ["graphql", "api", "advanced"]
  }) {
    id
    title
    updatedAt
  }
}
```

#### Delete Note
```graphql
mutation {
  deleteNote(id: "note_id")
}
```

### Search Features

#### Keyword Search
Basic text search in titles and content:
```graphql
query {
  searchNotes(query: "graphql") {
    id
    title
    content
    tags
  }
}
```

#### Search by Tag
```graphql
query {
  notesByTag(tag: "graphql") {
    id
    title
    tags
  }
}
```

#### Semantic Search (AI)
Find notes by meaning, not just keywords:
```graphql
query {
  semanticSearch(query: "API design patterns and best practices") {
    note {
      id
      title
      content
    }
    score
    reason
    matchType
  }
}
```

#### Hybrid Search (AI)
Combines keyword and semantic search:
```graphql
query {
  hybridSearch(query: "authentication security") {
    note {
      id
      title
      content
    }
    score
    matchType
  }
}
```

#### Smart Search (AI)
Advanced search with reasoning and gap detection:
```graphql
query {
  smartSearch(query: "How to optimize database queries?") {
    results {
      note {
        id
        title
        content
      }
      score
      reason
    }
    explanation
    gaps
    connections
  }
}
```

### Smart Linking System

The linking system enables you to create a knowledge graph by connecting related notes bidirectionally.

#### Create a Link Between Notes
```graphql
mutation {
  linkNotes(
    sourceId: "683b9e869e5b8965550ac605e8e035bf"
    targetId: "07be3a74f128355209eff2c4b0aa03f5"
    description: "Alternative approach to REST - compare design philosophies"
  ) {
    id
    sourceNoteId
    targetNoteId
    description
    createdAt
  }
}
```

#### View Note with All Connections
```graphql
query {
  note(id: "683b9e869e5b8965550ac605e8e035bf") {
    id
    title
    content
    
    # Outgoing links (notes this one points to)
    links {
      id
      targetNoteId
      description
      createdAt
    }
    
    # Incoming links (notes that point to this one)
    backlinks {
      id
      sourceNoteId
      description
      createdAt
    }
  }
}
```

#### Query Links Directly
Get all outgoing links from a note:
```graphql
query {
  links(noteId: "683b9e869e5b8965550ac605e8e035bf") {
    id
    targetNoteId
    description
    createdAt
  }
}
```

Get all backlinks (incoming links) to a note:
```graphql
query {
  backlinks(noteId: "683b9e869e5b8965550ac605e8e035bf") {
    id
    sourceNote {
      id
      title
      content
    }
    description
    createdAt
  }
}
```

#### Delete a Link
```graphql
mutation {
  unlinkNotes(linkId: "c0ec2f441ac251163543e12d9c6992c9")
}
```

### Link System Features

- **Bidirectional Tracking**: Automatically tracks both outgoing links and backlinks
- **Optional Descriptions**: Add context explaining the relationship between notes
- **Cascade Deletion**: Links automatically removed when notes are deleted (database constraint)
- **Self-Link Prevention**: System prevents linking a note to itself
- **Validation**: Both source and target notes must exist before creating a link
- **Field Resolvers**: Links and backlinks are lazily loaded only when requested

### Example: Building a Knowledge Graph

```graphql
# 1. Create notes about related topics
mutation {
  note1: createNote(input: {
    title: "Programming Fundamentals"
    content: "Variables, loops, functions"
    tags: ["programming", "basics"]
  }) { id }
  
  note2: createNote(input: {
    title: "Object-Oriented Programming"
    content: "Classes, objects, inheritance"
    tags: ["programming", "oop"]
  }) { id }
  
  note3: createNote(input: {
    title: "Design Patterns"
    content: "Common solutions to recurring problems"
    tags: ["programming", "architecture"]
  }) { id }
}

# 2. Link them together
mutation {
  link1: linkNotes(
    sourceId: "fundamentals_id"
    targetId: "oop_id"
    description: "OOP builds upon fundamental programming concepts"
  ) { id }
  
  link2: linkNotes(
    sourceId: "oop_id"
    targetId: "patterns_id"
    description: "Design patterns are best practices in OOP"
  ) { id }
}

# 3. Explore the knowledge graph
query {
  note(id: "oop_id") {
    title
    backlinks {
      sourceNoteId
      description
    }
    links {
      targetNoteId
      description
    }
  }
}
```

## Project Structure

```
graphql-pkm/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── gql/
│   │   ├── generated/
│   │   │   └── generated.go     # Generated GraphQL interfaces
│   │   ├── resolvers/
│   │   │   ├── resolver.go      # Resolver struct & registration
│   │   │   ├── query.go         # Query resolvers
│   │   │   ├── mutation.go      # Mutation resolvers
│   │   │   ├── note.go          # Note field resolvers (links/backlinks)
│   │   │   ├── search.go        # Search resolvers
│   │   │   └── subscription.go  # Subscription resolvers
│   │   └── schema.graphqls      # GraphQL schema definition
│   ├── ai/
│   │   └── client.go            # Groq AI client
│   ├── database/
│   │   ├── interface.go         # Database interface
│   │   ├── mysql.go             # MySQL implementation
│   │   └── memory.go            # In-memory fallback
│   ├── models/
│   │   └── models.go            # Domain models (Note, Link, etc.)
│   └── service/
│       ├── noteService.go       # Note business logic
│       ├── aiService.go         # AI operations
│       └── searchService.go     # Search logic
├── config/
│   └── config.go                # Configuration management
├── gqlgen.yml                   # gqlgen configuration
├── .env                         # Environment variables
├── go.mod                       # Go dependencies
└── README.md
```

## Development

### Adding New Features

1. **Update GraphQL schema** (`internal/gql/schema.graphqls`):
```graphql
type NewType {
  id: ID!
  field: String!
}

extend type Query {
  newQuery: NewType
}
```

2. **Regenerate GraphQL code**:
```bash
go run github.com/99designs/gqlgen generate
```

3. **Implement resolver** in appropriate file:
```go
// internal/gql/resolvers/query.go
func (r *queryResolver) NewQuery(ctx context.Context) (*models.NewType, error) {
    // Implementation
}
```

4. **Add service layer logic** if needed

5. **Test the new feature**

### Code Generation Best Practices

The project uses a **manual model + generated interface** approach:

**DO:**
- Define models in `internal/models/models.go`
- Let gqlgen generate only interfaces in `internal/gql/generated/`
- Split resolvers into logical files (query.go, mutation.go, etc.)
- Use field resolvers for computed or related data (like links/backlinks)

**DON'T:**
- ❌ Let gqlgen generate models (causes conflicts with database layer)
- ❌ Put all resolvers in one file (hard to maintain)
- ❌ Modify generated files manually (they'll be overwritten)

### gqlgen Configuration

Key parts of `gqlgen.yml`:

```yaml
# Use our custom models, don't generate them
autobind:
  - "graphql-pkm/internal/models"

models:
  Note:
    model: graphql-pkm/internal/models.Note
    fields:
      links:
        resolver: true      # Generate field resolver
      backlinks:
        resolver: true      # Generate field resolver
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/service/...

# Verbose output
go test -v ./...
```

### Building for Production

```bash
# Build binary
go build -o pkm-server cmd/server/main.go

# Run binary
./pkm-server

# Build with optimizations
go build -ldflags="-s -w" -o pkm-server cmd/server/main.go
```

## AI Features Deep Dive

### Embedding Cache

The system caches AI-generated embeddings to optimize performance and reduce costs:

#### Cache Statistics

Check cache performance in MySQL:
```sql
SELECT 
  COUNT(*) as total_cached,
  MIN(created_at) as oldest,
  MAX(created_at) as newest
FROM embeddings_cache;
```

#### Clear Cache

```sql
-- Clear all embeddings (useful when changing AI models)
DELETE FROM embeddings_cache;

-- Clear specific note
DELETE FROM embeddings_cache WHERE note_id = 'your_note_id';
```

## Database Schema

### Notes Table
```sql
CREATE TABLE notes (
  id VARCHAR(36) PRIMARY KEY,
  title TEXT NOT NULL,
  content LONGTEXT NOT NULL,
  tags JSON,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  summary TEXT,
  key_concepts JSON,
  INDEX idx_created_at (created_at),
  INDEX idx_updated_at (updated_at)
);
```

### Links Table
```sql
CREATE TABLE links (
  id VARCHAR(36) PRIMARY KEY,
  source_note_id VARCHAR(36) NOT NULL,
  target_note_id VARCHAR(36) NOT NULL,
  description TEXT,
  created_at DATETIME(6) NOT NULL,
  FOREIGN KEY (source_note_id) REFERENCES notes(id) ON DELETE CASCADE,
  FOREIGN KEY (target_note_id) REFERENCES notes(id) ON DELETE CASCADE,
  INDEX idx_source_note (source_note_id),
  INDEX idx_target_note (target_note_id)
);
```

### Embeddings Cache Table
```sql
CREATE TABLE embeddings_cache (
  note_id VARCHAR(36) PRIMARY KEY,
  embedding_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE
);
```

## Deployment

### Docker Deployment

**Dockerfile**:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o pkm-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/pkm-server .
COPY .env .

EXPOSE 8080
CMD ["./pkm-server"]
```

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: pkm
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENVIRONMENT=production
      - dbUrl=root:password@tcp(mysql:3306)/pkm?parseTime=true
    depends_on:
      - mysql

volumes:
  mysql_data:
```

**Deploy**:
```bash
docker-compose up -d
```

### Production Checklist

- Set `ENVIRONMENT=production`
- Use strong MySQL passwords
- Enable HTTPS/TLS
- Set up database backups
- Configure rate limiting
- Monitor API costs
- Set up logging
- Use connection pooling
- Enable database indices (auto-created)

## Performance Optimization

### Database Indexes

The system automatically creates indexes on:
- `notes.created_at`
- `notes.updated_at`
- `links.source_note_id`
- `links.target_note_id`
- `embeddings_cache.note_id`

### Query Optimization Tips

1. **Use specific fields** instead of fetching all data
2. **Leverage field resolvers** - links/backlinks are only loaded when requested
3. **Use the embedding cache** for repeated searches
4. **Batch operations** when creating multiple notes
5. **Use tags** for quick filtering
6. **Leverage database indexes** for fast lookups

## Troubleshooting

### Database Connection Issues

```bash
# Test MySQL connection
mysql -u root -p -h localhost

# Check if database exists
SHOW DATABASES;

# Verify tables
USE pkm;
SHOW TABLES;
DESCRIBE notes;
DESCRIBE links;
```

### Link-Related Issues

**Links not appearing in queries?**
- Ensure you have the `Note()` resolver method in `resolver.go`
- Check that `note.go` field resolvers exist
- Verify `links` and `backlinks` have `resolver: true` in `gqlgen.yml`
- Regenerate with: `go run github.com/99designs/gqlgen generate`

**Can't create links?**
- Both notes must exist before linking
- Check that both note IDs are correct
- Cannot link a note to itself
- Check MySQL foreign key constraints

### AI Features Not Working

1. Check API key is set: `echo $apiKey`
2. Get free Groq API key at: https://console.groq.com
3. Verify API endpoint is reachable
4. Check logs for API errors
5. Groq has generous free tier limits

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change PORT in .env
PORT=8081
```

### gqlgen Generation Errors

```bash
# Clean regeneration
rm internal/gql/generated/generated.go
go run github.com/99designs/gqlgen generate

# Verbose output
go run github.com/99designs/gqlgen generate --verbose

# Check for duplicate resolver methods
grep -r "func (r \*queryResolver)" internal/gql/resolvers/
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [gqlgen](https://github.com/99designs/gqlgen) - GraphQL server library for Go
- [Groq](https://groq.com) - Ultra-fast AI inference platform
- [MySQL](https://www.mysql.com) / [MariaDB](https://mariadb.org) - Reliable database systems
- Built with Go