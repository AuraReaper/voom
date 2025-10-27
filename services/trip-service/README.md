# Trip Service

A microservice for managing trips and fare calculations using Echo framework.

## Architecture

The service follows Clean Architecture principles with the following structure:

```
services/trip-service/
├── cmd/                    # Application entry points
│   └── main.go            # Main application setup with Echo server
├── internal/              # Private application code
│   ├── domain/           # Business domain models and interfaces
│   ├── service/          # Business logic implementation
│   ├── handlers/         # HTTP handlers (Echo)
│   ├── routes/           # Route definitions
│   └── infrastructure/   # External dependencies implementations
│       ├── events/       # Event handling (RabbitMQ)
│       ├── grpc/         # gRPC server handlers
│       └── repository/   # Data persistence
├── pkg/                  # Public packages
│   └── types/           # Shared types and models
└── README.md            # This file
```

## API Endpoints

### Health Check
- **GET** `/health` - Check service health

### Trips
- **POST** `/api/v1/trips` - Create a new trip
- **GET** `/api/v1/trips/:id` - Get trip by ID
- **GET** `/api/v1/trips?user_id=<user_id>` - List trips (optionally filter by user)
- **PUT** `/api/v1/trips/:id` - Update trip
- **DELETE** `/api/v1/trips/:id` - Delete trip

### Fares
- **POST** `/api/v1/fares/calculate` - Calculate fare for a trip

## Example Requests

### Create Trip
```bash
curl -X POST http://localhost:8080/api/v1/trips \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "package_type": "van",
    "total_price": 25.50
  }'
```

### Calculate Fare
```bash
curl -X POST http://localhost:8080/api/v1/fares/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "package_type": "van",
    "distance": 10.5,
    "duration": 30
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Running the Service

```bash
# From project root
go run services/trip-service/cmd/main.go
```

The service will start on port 8080.

## Package Types
- `bike` - Rate: $1.5/km
- `car` - Rate: $2.0/km  
- `van` - Rate: $3.0/km
- Base fare: $5.0 for all types

## Key Benefits

1. **Echo Framework**: High-performance HTTP router and middleware
2. **Clean Architecture**: Clear separation of concerns
3. **RESTful API**: Standard HTTP methods and status codes
4. **JSON API**: Easy integration with frontend applications
5. **Middleware Support**: Logging, CORS, recovery, and custom middleware
