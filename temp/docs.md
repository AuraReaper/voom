# Voom: The Technical "Holy Bible"

Welcome to the definitive guide to Voom. This document is designed to give you a mastery-level understanding of the project, explaining not just *what* was built, but the deep technical *why* behind every decision.

---

## 1. Core Philosophy: Why this Architecture?

Voom isn't just a ride-sharing app; it's a demonstration of a **highly resilient, scalable distributed system**.

### Why Microservices?
- **Independent Scaling**: The `Driver Service` (high write-volume for locations) needs different resources than the `Payment Service` (high security, low volume).
- **Fault Isolation**: If the `Payment Service` goes down, users can still book rides; we just process the transaction later via RabbitMQ retries.
- **Polyglot Potential**: While we use Go now, the microservice boundary allows us to plug in a Python-based ML service for pricing in the future without touching the core.

### Why Go?
- **Concurrency**: Go's goroutines make handling thousands of concurrent WebSocket connections and gRPC streams extremely efficient with minimal memory overhead.
- **Type Safety**: Protobuf + Go ensures that our architectural "contracts" are strictly enforced.

---

## 2. Service-by-Service Deep Dive

### 2.1 API Gateway (The Entry Point)
Built with the **Echo** framework, the Gateway performs three critical roles:

1.  **Request Routing & Authentication**: It acts as the gatekeeper, ensuring internal services are never exposed directly to the internet.
2.  **WebSocket Connection Management**: 
    - It uses a custom `ConnectionManager` (in `shared/messaging`) to track active user sessions.
    - **Implementation Detail**: We use `connManager.Upgrade` to turn HTTP requests into persistent WebSocket connections for real-time rider and driver streams.
3.  **Cross-Service Translation**: It translates incoming JSON/REST requests from the frontend into efficient gRPC calls for internal services.

### 2.2 Trip Service (The Orchestrator)
The brain of the system. It manages the trip state machine: `IDLE` -> `PREVIEW` -> `CREATED` -> `ASSIGNED` -> `STARTED` -> `COMPLETED`.

- **OSRM Integration**: We use the **Open Source Routing Machine (OSRM)** API.
    - **Why?** It gives us precise distance and duration data for route geometry, which is essential for accurate pricing and drawing the route on the frontend map.
- **Dynamic Pricing Algorithm**:
    - Pricing is calculated in `internal/service/service.go`.
    - It combines a `BaseFare` (based on vehicle type) + `DistanceFare` (`distanceKm * pricePerKm`) + `TimeFare` (`durationMin * pricePerMinute`).
    - This allows us to adjust prices dynamically based on real-world traffic data from OSRM.

### 2.3 Driver Service (Location & Matching)
The most data-intensive service.

- **Geohashing Logic**: 
    - We use the `mmcloughlin/geohash` library. 
    - **The Why**: Standard relational database spatial queries can be slow. By converting `(lat, lon)` into a 1D string (geohash), we can use MongoDB's standard prefix indexing to find drivers in a specific area in $O(log N)$ time.
    - This allows us to query "all drivers whose location starts with 'u4p'" to instantly find nearby cars.

### 2.4 Payment Service (Decoupled Reliability)
Handles the money.

- **Stripe Integration**: Uses the official Stripe Go SDK.
- **Asynchronous Flow**: When a trip is created, a command is sent via RabbitMQ to the Payment Service. This decoupling ensures that even if Stripe's API is slow, the user doesn't feel the lag in the "Ride Request" flow.

---

## 3. The Messaging Backbone (RabbitMQ)

We use **RabbitMQ with Topic Exchanges** for maximum flexibility.

### Reliability Patterns
1.  **Manual Acknowledgments (ACKs)**: We never use auto-ack. A message is only removed from the queue `if err := handler(ctx, d); err == nil`. This prevents data loss during service crashes.
2.  **Exponential Backoff**: In `shared/retry`, we implemented a strategy where failed tasks (like a Stripe API timeout) are retried with increasing delays (1s, 2s, 4s...) to avoid overwhelming downstream services.
3.  **Persistent Delivery**: Every message is marked with `amqp.Persistent` so they survive a RabbitMQ server restart.

---

## 4. Frontend Architecture (Next.js & Maps)

The frontend isn't just a UI; it's a real-time tracking dashboard.

- **WebSocket Hooks**: We built custom hooks like `useRiderStreamConnection` and `useDriverStreamConnection`.
    - **Why hooks?** This encapsulates the connection logic, auto-reconnects, and event parsing, keeping the UI components clean and focused on rendering.
- **Map Visualization**: Using **Leaflet** and **React-Leaflet**.
    - We map the `OSRM Geometry` (coordinates) directly onto the map to draw the blue route line.
    - Driver locations are pushed via WebSockets and updated in the React state, causing the car markers to move smoothly without a page refresh.

---

## 5. Observability: "The Doctor Service"

In a microservice world, debugging is hard. We solve this with **OpenTelemetry**.

- **TraceID Propagation**: Every request (REST or gRPC) and every message (RabbitMQ) carries a `TraceID` in its header.
- **The Payoff**: You can look at a single trace and see: 
    1. Request hits Gateway.
    2. Gateway calls Trip Service.
    3. Trip Service publishes a RabbitMQ event.
    4. Driver Service consumes that event.
    If the car doesn't show up, you can see exactly which step failed.

---

## 6. Advanced Q&A: The "Griller" Preparation

**Q: "What happens if two drivers accept the same trip simultaneously?"**
*A: I use **Optimistic Locking** in the database. The first driver to update the Trip record status from `pending` to `assigned` wins. The second driver's update will fail the version check, and they'll receive a 'Too Late' notification.*

**Q: "How do you handle 'Gaps' in RabbitMQ message delivery order?"**
*A: Since we use RabbitMQ Topic Exchanges, order isn't strictly guaranteed for different users. However, for a single trip, we use Idempotency Keys (like TripID). If a 'TripCompleted' message arrives before 'TripStarted' (unlikely but possible), the service layer checks the current state in the DB and discards the invalid state transition.*

**Q: "Why MongoDB instead of Postgres?"**
*A: Write-heavy location updates. MongoDB's ability to handle high insert/update volume for real-time driver coordinates is superior to standard relational databases without complex partitioning. Also, the flexible schema allows us to store varying route geometries easily.*
