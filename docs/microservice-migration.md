# Microservice Migration

Each module is designed to be self-contained and can be easily extracted into a microservice.

## Migration Steps

1. **Copy module directory** to new service repository
2. **Update dependencies**: Replace shared infrastructure with service-specific connections
3. **Add inter-service communication**: Implement gRPC or REST clients for cross-service calls
4. **Update shared components**: Either copy shared components or use a shared library
5. **No refactoring needed**: Module structure remains the same

## Example: Extracting Leaderboard Module

```
leaderboard-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   ├── shared/          # Copy or use shared library
│   │   └── redis/       # Redis connection (shared infrastructure)
│   └── module/
│       └── leaderboard/ # Copy entire module - no changes needed
```

## Benefits

- **Clear Boundaries**: Each module owns its domain, infrastructure, and presentation
- **Easy Testing**: Mock repository interfaces at domain level
- **Independent Deployment**: Modules can be versioned and deployed separately
- **Team Ownership**: Different teams can own different modules
- **Technology Flexibility**: Each module can use different technologies

