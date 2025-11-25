# GraphQL Personal Knowledge Management (PKM) System

A GraphQL-based Personal Knowledge Management system with optional AI-powered search, built using Go and OpenRouter.

## Features

- Note management (create, read, update, delete)
- Smart linking with backlink tracking
- AI-assisted semantic and hybrid search
- Tagging system
- Real-time GraphQL API
- AI-based insight and connection suggestions

## Tech Stack

- **Backend**: Go 1.21+
- **GraphQL**: gqlgen
- **AI**: OpenRouter API integration
- **Database**: In-memory (extendable to PostgreSQL)
- **Configuration**: godotenv

## Installation

### Requirements
- Go 1.21 or newer  
- OpenRouter API key (optional, for AI features)

### Steps

1. Clone repository:
```bash
git clone <repository>
cd learn-graphql-pkm
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment:
```bash
cp .env.example .env
```

4. Generate GraphQL code:
```bash
go run github.com/99designs/gqlgen generate
```

5. Run server:
```bash
go run cmd/server/main.go
```

6. Open GraphQL Playground:  
http://localhost:8080

## Configuration

`.env` file variables:

```bash
PORT=8080
Environment=development
apiKey=your_openrouter_api_key_here
apiUrl=https://openrouter.ai/api/v1
defaultModel=deepseek/deepseek-coder:33b-instruct
```

### Supported AI Models

- deepseek/deepseek-coder:33b-instruct  
- deepseek/deepseek-chat  
- anthropic/claude-3-sonnet:beta  
- openai/gpt-3.5-turbo  
- openai/gpt-4  

## API Usage

### Create Note
```graphql
mutation {
  createNote(input: {
    title: "GraphQL Basics",
    content: "GraphQL is a query language for APIs",
    tags: ["graphql", "api", "learning"]
  }) {
    id
    title
    tags
    createdAt
  }
}
```

### Get Notes
```graphql
query {
  notes {
    id
    title
    content
    tags
    createdAt
  }
}
```

### Keyword Search
```graphql
query {
  searchNotes(query: "graphql") {
    id
    title
    content
  }
}
```

### Semantic Search
```graphql
query {
  semanticSearch(query: "API design patterns") {
    note {
      title
      content
    }
    score
    reason
    matchType
  }
}
```

### Smart Search
```graphql
query {
  smartSearch(query: "How to optimize database queries") {
    results {
      note {
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

### Hybrid Search
```graphql
query {
  hybridSearch(query: "authentication best practices") {
    note {
      title
      content
    }
    score
    matchType
  }
}
```

### Update Note
```graphql
mutation {
  updateNote(id: "note_id", input: {
    title: "Updated Title",
    content: "Updated content",
    tags: ["updated", "tags"]
  }) {
    id
    title
    updatedAt
  }
}
```

### Delete Note
```graphql
mutation {
  deleteNote(id: "note_id")
}
```

## Project Structure

```
learn-graphql-pkm/
├── cmd/server/main.go
├── internal/
│   ├── gql/
│   ├── ai/
│   ├── database/
│   ├── models/
│   └── service/
├── config/
└── go.mod
```

## Development

### Adding Features

1. Update schema (`internal/gql/schema.graphqls`)
2. Add resolvers
3. Add service logic
4. Regenerate GraphQL code:
```bash
go run github.com/99designs/gqlgen generate
```

### Testing
```bash
go test ./...
go test -cover ./...
```

### Building
```bash
go build -o pkm-server cmd/server/main.go
```

## AI Features

- Semantic search with embeddings  
- Smart search with reasoning  
- Gap detection  
- Suggested note connections  
- Embedding caching to reduce cost  

## Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/server/main.go
EXPOSE 8080
CMD ["./main"]
```

### Production Environment Variables
```bash
PORT=8080
Environment=production
apiKey=your_production_api_key
apiUrl=https://openrouter.ai/api/v1
defaultModel=deepseek/deepseek-coder:33b-instruct
```

## License

MIT License
