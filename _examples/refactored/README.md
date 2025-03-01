# Refactored Router Example

This example demonstrates the refactored router architecture which separates concerns into distinct packages:

## Package Structure

- `router`: Core routing functionality
- `docs`: API documentation utilities
- `metadata`: Shared type definitions
- `openapi`: OpenAPI specification generation
- `swagger`: Swagger UI configuration and serving
- `integration`: Adapters connecting the components

## Features Demonstrated

1. Clean separation of routing and documentation
2. Type-safe route documentation using the `docs` package
3. Shared metadata types for consistency
4. Optional OpenAPI and Swagger UI integration
5. Modern Go practices (generics, reflection)

## Usage

The example shows:
- Defining routes with documentation
- Configuring OpenAPI generation
- Setting up Swagger UI
- Using type-safe request/response documentation
- Validation tag support

To run:

```bash
go run main.go
```

Then visit:
- API: http://localhost:8080/users
- Documentation: http://localhost:8080/docs
- OpenAPI Spec: http://localhost:8080/openapi.json

## Benefits of the Refactoring

1. **Modularity**: Each package has a single responsibility
2. **Optional Features**: Use only what you need
3. **Type Safety**: Better compile-time guarantees
4. **Maintainability**: Clear separation of concerns
5. **Extensibility**: Easy to add new integrations

## Example Routes

- `GET /users`: List all users
- `POST /users`: Create a new user
- `GET /users/:id`: Get user by ID

Each route is documented using the new `docs` package, which provides a clean and type-safe way to document your API.