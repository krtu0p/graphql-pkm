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
- **Database**: MySQL (with in-memory fallback option)
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
    keyTopics
    links {
      id
      targetNote {
        title
      }
    }
    backlinks {
      id
      sourceNote {
        title
      }
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

#### Create a Link Between Notes
```graphql
mutation {
  linkNotes(
    sourceId: "note_abc123"
    targetId: "note_def456"
    description: "Related concept - prerequisite knowledge"
  ) {
    id
    sourceNoteId
    targetNoteId
    description
    createdAt
  }
}
```

#### Get Outgoing Links
Get all links where this note is the source:
```graphql
query {
  noteLinks(noteId: "note_abc123") {
    id
    description
    targetNote {
      id
      title
      content
    }
  }
}
```

#### Get Backlinks (Incoming Links)
Get all links pointing to this note:
```graphql
query {
  noteBacklinks(noteId: "note_abc123") {
    id
    description
    sourceNote {
      id
      title
      content
    }
  }
}
```

#### Get Note with All Connections
```graphql
query {
  note(id: "note_abc123") {
    id
    title
    content
    
    # Outgoing links
    links {
      description
      targetNote {
        id
        title
      }
    }
    
    # Incoming links (backlinks)
    backlinks {
      description
      sourceNote {
        id
        title
      }
    }
  }
}
```

#### Delete a Link
```graphql
mutation {
  unlinkNotes(linkId: "link_xyz789")
}
```

### Link Features

- **Bidirectional tracking**: Automatically track both outgoing links and backlinks
- **Optional descriptions**: Add context to explain relationships
- **Cascade deletion**: Links automatically removed when notes are deleted
- **Self-link prevention**: System prevents linking a note to itself
- **Validation**: Both source and target notes must exist

## Project Structure

```
graphql-pkm/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── gql/
│   │   ├── generated/           # Generated GraphQL code
│   │   ├── models/              # GraphQL models
│   │   ├── resolvers/           # Query/Mutation resolvers
│   │   └── schema.graphqls      # GraphQL schema definition
│   ├── ai/
│   │   └── client.go            # OpenRouter AI client
│   ├── database/
│   │   ├── interface.go         # Database interface
│   │   ├── mysql.go             # MySQL implementation
│   │   ├── memory.go            # In-memory implementation
│   │   └── embeddings_cache.go  # Embedding cache
│   ├── models/
│   │   └── models.go            # Domain models
│   └── service/
│       ├── noteService.go       # Note business logic
│       ├── aiService.go         # AI operations
│       └── searchService.go     # Search logic
├── config/
│   └── config.go                # Configuration management
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

3. **Implement resolver**:
```go
func (r *queryResolver) NewQuery(ctx context.Context) (*models.NewType, error) {
    // Implementation
}
```

4. **Add service layer logic** if needed

5. **Test the new feature**

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

### Environment Variables (Production)

```bash
PORT=8080
ENVIRONMENT=production
apiKey=gsk_your_production_groq_key
apiUrl=https://api.groq.com/openai/v1
defaultModel=llama-3.3-70b-versatile
dbUrl=user:secure_password@tcp(db-host:3306)/pkm?parseTime=true
```

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
2. **Leverage the cache** by searching similar queries
3. **Batch operations** when creating multiple notes
4. **Use tags** for quick filtering

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
```

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

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

-  Email: support@example.com
-  Issues: [GitHub Issues](https://github.com/yourusername/graphql-pkm/issues)

## Acknowledgments

- [gqlgen](https://github.com/99designs/gqlgen) - GraphQL server library
- [Groq](https://groq.com) - Ultra-fast AI inference platform
- Built using Go