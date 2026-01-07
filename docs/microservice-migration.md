# Microservice Migration

Each module is designed to be self-contained and can be easily extracted into a microservice.

## Migration Steps

1. **Copy module content** (all clean architecture layers: domain, application, adapters, infrastructure) to new service repository
2. **Remove module wrapper**: Place layers directly under `internal/` (no `module/` directory in microservices)
3. **Remove shared infrastructure**: Each microservice manages its own infrastructure connections (no `shared/` directory)
4. **Update dependencies**: Replace shared infrastructure with service-specific connections
5. **Add inter-service communication**: Implement gRPC or REST clients for cross-service calls
6. **No refactoring needed**: Clean architecture layer structure remains the same

## Example: Extracting Leaderboard Module

```
leaderboard-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── domain/          # Domain entities and repository interfaces
│   ├── application/     # Use cases and business logic
│   ├── adapters/        # HTTP/WebSocket handlers
│   └── infrastructure/  # Repository implementations (includes Redis, DB connections)
```

## Benefits

- **Clear Boundaries**: Each module owns its domain, infrastructure, and presentation
- **Easy Testing**: Mock repository interfaces at domain level
- **Independent Deployment**: Modules can be versioned and deployed separately
- **Team Ownership**: Different teams can own different modules
- **Technology Flexibility**: Each module can use different technologies

