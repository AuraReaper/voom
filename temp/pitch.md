# Voom: Project Pitch Script

*Use this script for a 1-2 minute introduction when asked, "Tell me about a project you've worked on."*

---

"I’d love to tell you about **Voom**, a high-performance ride-sharing platform I built from the ground up using a **Go-based microservices architecture**.

The core mission was to create a scalable system capable of handling real-time driver tracking, complex trip lifecycles, and secure payments. 

Technically, the project is structured into four main services:
1.  An **API Gateway** that handles high-concurrency requests using the Echo framework.
2.  A **Trip Service** that orchestrates the entire ride flow—from fare estimation to completion.
3.  A **Driver Service** that uses **Geohashing algorithms** for efficient proximity-based driver searching.
4.  And a **Payment Service** integrated with **Stripe** for secure transactions.

One of the most interesting challenges was managing real-time updates. I used **gRPC** for low-latency inter-service communication and **WebSockets** to push live driver location updates to the frontend. I also implemented an asynchronous event-driven flow using **RabbitMQ** to ensure the system remains resilient even during peak loads.

On the frontend, I used **Next.js** with **Leaflet** for interactive maps and **Tailwind CSS** for a premium UI. To ensure high observability, I integrated **OpenTelemetry** for distributed tracing across all services.

The project really pushed me to think about system design, especially around data consistency and latency in a distributed environment."
