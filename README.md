# Coupon System

## Description

This project implements a backend service for managing and applying coupons. It provides APIs for creating coupons, validating coupons against user carts, and fetching applicable coupons based on user and cart details. The system incorporates features like various discount types, usage restrictions, time validity, and integration with medicine and category IDs for targeted promotions.

## Architecture

The project follows a layered architecture:

- **API Layer (`internal/api/handlers`):** Handles incoming HTTP requests, parses request bodies, and calls the appropriate service methods. It also formats service responses for HTTP output.
- **Services Layer (`internal/services`):** Contains the core business logic. It orchestrates operations by interacting with the data layer and applying complex validation and calculation rules.
- **Data Layer (`internal/storage/database`):** Abstracts database interactions. It provides an interface for performing CRUD operations on coupon data. The current implementation uses SQLite.
- **Models Layer (`internal/models`):** Defines the data structures used throughout the application for requests, responses, and database entities.
- **Caching Layer (`internal/caching`):** Provides an interface and implementation for caching data, currently used for caching applicable coupon results.

The API handlers interact with the services layer, and the services layer interacts with the storage and caching layers.

## Setup Instructions

**Prerequisites:**

- Go (version 1.24 or later recommended)

**Steps:**

1.  **Clone the repository:**
    
```bash
git clone <repository_url>
```

2. **Seed Database:**

```bash
go run ./cmd/seed/main.go
```

3. **Setup Environment Secret:**

```bash
export JWT_SECRET=<secret>
export CGO_ENABLED=1
```

4. **Run the project:**
    
```bash
go run ./cmd/coupon_server/main.go
```

    By default, the application will use a SQLite database file named `coupons.db` in the current directory. If you want to change this behavior update internal/config/config.go

## Concurrency, Caching, and Locking

- **Concurrency:** Go's goroutines and channels are leveraged for handling concurrent requests efficiently within the API and service layers. The database interactions are handled by the GORM library, which manages database connection pooling to handle concurrent database access.
- **Caching:** The `GetApplicableCoupons` method in the service layer utilizes a caching mechanism (`internal/caching`) to store and retrieve results based on the user ID and request parameters. This reduces the load on the database for frequently requested applicable coupon lists. The current implementation uses an in-memory LRU cache. Cache invalidation is not explicitly implemented and would require a strategy based on data changes.
- **Locking:** The current implementation primarily relies on the underlying database's transaction and locking mechanisms for ensuring data consistency during operations like updating coupon usage. No explicit application-level locking is implemented within the core coupon logic, but it might be necessary for more complex scenarios involving shared resources beyond the database.

## API Documentation

The API documentation is available in Swagger format.

- [Swagger JSON](docs/swagger.json)
- [Swagger YAML](docs/swagger.yaml)

## Deployed

Swagger UI - https://typical-teddy-darshandrander-52b1eff3.koyeb.app/swagger/index.html

## Future implementations

- Carts
- User management