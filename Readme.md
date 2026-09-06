# Mini-PaaS

 A small self-hosted Platform-as-a-Service (PaaS) built with **Go, PostgreSQL, RabbitMQ, and Docker**.

 The goal of this project is to understand how a simplified deployment platform works internally — from receiving a GitHub webhook to building an application, running it inside Docker, and exposing it through a reverse proxy.

 > **Status:** 🚧 Work in Progress

---

 ## Architecture

```
                    ┌─────────────────────┐
                    │       GitHub        │
                    └──────────┬──────────┘
                               │
                            webhook
                               │
                               ▼
                    ┌─────────────────────┐
                    │    Go Control Plane │
                    │                     │
                    │  API / Projects     │
                    │  Deployments        │
                    │  Build Management   │
                    └──────────┬──────────┘
                               │
                        deployment job
                               │
                               ▼
                    ┌─────────────────────┐
                    │       Queue         │
                    │      RabbitMQ       │
                    └──────────┬──────────┘
                               │
                         consume job
                               │
                               ▼
                    ┌─────────────────────┐
                    │      Go Worker      │
                    │                     │
                    │  Clone repository   │
                    │  Build image        │
                    │  Start container    │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │       Docker        │
                    │                     │
                    │  app-1 : 3001       │
                    │  app-2 : 3002       │
                    │  app-3 : 3003       │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │    Reverse Proxy    │
                    │         Go          │
                    └──────────┬──────────┘
                               │
                               ▼
                       app1.localhost
                       app2.localhost
```

---

 ## What is a PaaS?

 A Platform-as-a-Service abstracts away the infrastructure required to deploy an application.

 Instead of manually doing:

```
git clone
docker build
docker run
configure networking
configure ports
configure routing
```

 a user should eventually be able to:

```
Connect GitHub repository
        ↓
Push code
        ↓
GitHub webhook
        ↓
Create deployment
        ↓
Build application
        ↓
Start container
        ↓
Application becomes available
```

 This project is a simplified implementation of that idea.

---

 ## Project Goals

 The main goal is not to build a production-ready Heroku or Railway clone.

 The goal is to understand the engineering concepts behind a deployment platform:

 - REST API design
- Project and deployment management
- PostgreSQL data modeling
- Foreign-key relationships
- Deployment state machines
- Background jobs
- RabbitMQ queues
- Worker architecture
- Docker image building
- Docker container lifecycle
- Dynamic port allocation
- Reverse proxying
- GitHub webhooks
- Graceful shutdown
- Service/repository architecture
- Eventually deployment logs and build history

---

 ## Core Components

 ### 1\. Go Control Plane

 The control plane is responsible for managing the state of the platform.

 It exposes the API used by users and communicates with the database and job queue.

 Responsibilities include:

 - User management
- Authentication
- Project management
- Deployment creation
- Deployment state management
- Build management
- GitHub webhook handling
- Creating deployment jobs

 The control plane does **not** perform the actual application build itself.

 Instead, it creates a deployment job and sends it to RabbitMQ.

```
HTTP Request
     │
     ▼
Go API
     │
     ├── PostgreSQL
     │
     └── RabbitMQ
             │
             ▼
           Worker
```

---

 ### 2\. PostgreSQL

 PostgreSQL stores the persistent state of the platform.

 The database contains information such as:

 - Users
- Projects
- Deployments
- Deployment status
- Repository information
- Timestamps
- Eventually builds and logs

 The basic relationship is:

```
User
 │
 └── Projects
       │
       ├── Deployment
       ├── Deployment
       └── Deployment
```

 A deployment belongs to a project through a foreign key:

```
Deployment.project_id
        │
        ▼
Project.id
```

 In Go/GORM this is represented by:

```
type Deployment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null"`
	Project   Project   `gorm:"foreignKey:ProjectID"`

	Status    DeploymentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

 This allows the deployment to access its associated project:

```
deployment.Project.Name
deployment.Project.RepositoryURL
```

 when loaded with GORM:

```
db.Preload("Project").First(&deployment, deploymentID)
```

---

 ## Deployment State Machine

 Deployments have a defined lifecycle.

 A deployment cannot arbitrarily jump between states.

```
             ┌──────────┐
             │  QUEUED  │
             └────┬─────┘
                  │
                  ▼
             ┌──────────┐
             │ BUILDING │
             └────┬─────┘
                  │
             ┌────┴─────┐
             │          │
             ▼          ▼
        ┌─────────┐  ┌─────────┐
        │ RUNNING │  │ FAILED  │
        └────┬────┘  └─────────┘
             │
        ┌────┴─────┐
        │          │
        ▼          ▼
   ┌─────────┐  ┌─────────┐
   │ STOPPED │  │ FAILED  │
   └────┬────┘  └─────────┘
        │
        │ restart
        ▼
   ┌─────────┐
   │ RUNNING │
   └─────────┘
```

 The transition rules are implemented inside the domain model:

```
func (d *Deployment) TransitionTo(
	status DeploymentStatus,
) error
```

 For example:

```
QUEUED   → BUILDING
BUILDING → RUNNING
BUILDING → FAILED
RUNNING  → STOPPED
RUNNING  → FAILED
STOPPED  → RUNNING
STOPPED  → FAILED
```

 Invalid transitions are rejected.

 For example:

```
QUEUED → RUNNING
```

 is invalid because a deployment must go through the build stage first.

 The state machine keeps the deployment lifecycle rules inside the domain model rather than spreading them across HTTP handlers.

---

 ## RabbitMQ

 RabbitMQ is used to decouple the API from the deployment worker.

 When a deployment is requested:

```
Client
  │
  ▼
Control Plane
  │
  ├── Create Deployment
  │       status = QUEUED
  │
  └── Publish Job
          │
          ▼
       RabbitMQ
```

 The API does not need to wait for the deployment to finish.

 The worker consumes the job asynchronously:

```
RabbitMQ
    │
    ▼
Go Worker
    │
    ├── Clone repository
    ├── Build Docker image
    ├── Start container
    └── Update deployment status
```

 This allows the control plane and workers to scale independently.

---

 ## Go Worker

 The worker is responsible for executing deployment jobs.

 A simplified deployment job might contain:

```
{
  "deployment_id": "uuid",
  "project_id": "uuid",
  "repository_url": "https://github.com/example/app",
  "commit_sha": "abc123"
}
```

 The worker then performs approximately:

```
Receive job
    │
    ▼
Clone repository
    │
    ▼
Checkout commit
    │
    ▼
Build Docker image
    │
    ▼
Start container
    │
    ▼
Assign port
    │
    ▼
Update deployment
    │
    ▼
RUNNING
```

 If something fails:

```
BUILDING
    │
    ▼
FAILED
```

---

 ## Docker

 Applications are isolated using Docker containers.

 For example:

```
Docker
│
├── app-1
│    └── host port 3001
│
├── app-2
│    └── host port 3002
│
└── app-3
     └── host port 3003
```

 The worker is responsible for:

 1. Building the image.
2. Creating the container.
3. Assigning a port.
4. Starting the container.
5. Tracking the container.
6. Stopping/removing it when required.

 The exact container naming and port allocation strategy will evolve as the project develops.

---

 ## Reverse Proxy

 The reverse proxy is responsible for routing incoming requests to the correct application container.

 For example:

```
app1.localhost
      │
      ▼
Reverse Proxy
      │
      ▼
localhost:3001
      │
      ▼
app-1 container
```

 And:

```
app2.localhost
      │
      ▼
Reverse Proxy
      │
      ▼
localhost:3002
      │
      ▼
app-2 container
```

 The proxy should eventually dynamically discover the correct deployment and route traffic based on the hostname.

 Conceptually:

```
Host header
     │
     ▼
app1.localhost
     │
     ▼
Deployment lookup
     │
     ▼
container port
     │
     ▼
Docker container
```

---

 # API

 ## Authentication

 ### Register

```
POST /auth/register
```

 ### Login

```
POST /auth/login
```

---

 ## Projects

 Projects are the top-level resource for applications deployed through the platform.

 ### Create project

```
POST /projects
```

 ### List projects

```
GET /projects
```

 ### Get project

```
GET /projects/:projectID
```

 ### Delete project

```
DELETE /projects/:projectID
```

---

 ## Deployments

 Deployments are associated with a project.

 Creating and listing deployments uses the project as the parent resource.

 ### Create deployment

```
POST /projects/:projectID/deployments
```

 Example:

```
POST /projects/8e4.../deployments
```

 The deployment handler receives the `projectID` and passes it to the deployment service.

 The deployment service creates:

```
Deployment
├── ID
├── ProjectID
├── Status = QUEUED
├── CreatedAt
└── UpdatedAt
```

 ### List project's deployments

```
GET /projects/:projectID/deployments
```

 ### Get deployment

 Once a deployment has its own ID, it becomes an independent resource:

```
GET /deployments/:deploymentID
```

 ### Update deployment status

```
PUT /deployments/:deploymentID/status
```

 Status transitions are validated by the deployment domain model.

---

 # Routing Philosophy

 The API intentionally uses both nested and top-level deployment routes.

```
/projects/:projectID/deployments
```

 is used when the operation is about deployments **belonging to a specific project**.

```
/deployments/:deploymentID
```

 is used when the deployment itself is the resource being operated on.

 For example:

```
POST /projects/:projectID/deployments
GET  /projects/:projectID/deployments

GET  /deployments/:deploymentID
PUT  /deployments/:deploymentID/status
```

 The URL structure does not determine which service owns the logic.

 Deployment operations are handled by:

```
DeploymentHandler
        ↓
DeploymentService
        ↓
DeploymentRepository
```

 even when the route is nested under `/projects`.

 This keeps `ProjectService` focused on project-related operations and prevents it from becoming responsible for the entire deployment lifecycle.

---

 # Application Structure

 The current Go application follows a simple layered architecture:

```
internal/
│
├── config/
│
├── database/
│
├── handlers/
│   ├── user_handler.go
│   ├── project_handler.go
│   └── deployment_handler.go
│
├── middlewares/
│   └── auth.go
│
├── models/
│   ├── user.go
│   ├── project.go
│   └── deployment.go
│
├── repository/
│   ├── user_repository.go
│   ├── project_repository.go
│   └── deployment_repository.go
│
└── service/
    ├── user_service.go
    ├── project_service.go
    └── deployment_service.go
```

 The general request flow is:

```
HTTP
 │
 ▼
Handler
 │
 ▼
Service
 │
 ▼
Repository
 │
 ▼
PostgreSQL
```

 For asynchronous deployment execution:

```
HTTP
 │
 ▼
Deployment Handler
 │
 ▼
Deployment Service
 │
 ├──────────────► PostgreSQL
 │
 └──────────────► RabbitMQ
                         │
                         ▼
                       Worker
                         │
                         ▼
                       Docker
```

---

 # Responsibilities

 ## Handler

 The handler is responsible for HTTP concerns:

 - Reading path parameters
- Reading request bodies
- Validating HTTP input
- Calling services
- Returning HTTP responses

 It should not contain deployment business logic.

---

 ## Service

 The service contains application/business logic.

 For example:

```
DeploymentService
├── CreateDeployment
├── GetDeployment
├── GetDeploymentsForProject
├── UpdateStatus
└── ...
```

 The service decides what should happen.

---

 ## Repository

 Repositories handle database persistence.

 For example:

```
DeploymentRepository
├── Create
├── GetByID
├── GetByProjectID
├── Update
└── ...
```

 The repository should not decide whether:

```
QUEUED → RUNNING
```

 is a valid transition.

 That rule belongs to the deployment domain model.

---

 # Authentication

 Protected application routes use authentication middleware:

```
protected := router.Group("/")
protected.Use(middlewares.AuthMiddleware())
```

 The current API structure is:

```
Public
│
├── GET  /health
├── POST /auth/register
└── POST /auth/login

Protected
│
├── /projects
│
└── /deployments
```

 Authentication will eventually be used not only to identify users, but also to enforce ownership:

```
User A
  │
  └── Project A
        │
        └── Deployment A
```

 User A should not be able to access:

```
User B
  └── Project B
        └── Deployment B
```

 even if User A knows the UUID.

---

 # Deployment Flow

 A complete deployment will eventually look like this:

```
1. User creates project
        │
        ▼
2. Project stored in PostgreSQL
        │
        ▼
3. GitHub repository connected
        │
        ▼
4. GitHub sends webhook
        │
        ▼
5. Control plane receives webhook
        │
        ▼
6. Deployment created
        │
        ▼
7. Status = QUEUED
        │
        ▼
8. Deployment job published to RabbitMQ
        │
        ▼
9. Worker consumes job
        │
        ▼
10. Status = BUILDING
        │
        ▼
11. Worker clones repository
        │
        ▼
12. Docker image is built
        │
        ▼
13. Docker container starts
        │
        ▼
14. Status = RUNNING
        │
        ▼
15. Reverse proxy routes traffic
        │
        ▼
16. app1.localhost
```

---

 # Failure Flow

 Failures should update the deployment state rather than crashing the entire control plane.

 For example:

```
QUEUED
   │
   ▼
BUILDING
   │
   │ build failed
   ▼
FAILED
```

 Or:

```
RUNNING
   │
   │ container crashed
   ▼
FAILED
```

 The deployment record remains available so users can inspect what happened.

 Eventually this will be extended with build logs and error information.

---

 # Local Development

 ## Requirements

 You will need:

 - Go
- PostgreSQL
- RabbitMQ
- Docker
- Git

 Verify the installations:

```
go version
docker --version
git --version
```

---

 ## Environment Variables

 Example:

```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/minipaas
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

 Additional configuration will be added as the project grows.

---

 # Running the Infrastructure

 Start PostgreSQL and RabbitMQ locally.

 Docker can be used to run the infrastructure dependencies.

 Example:

```
docker compose up -d
```

 Check running containers:

```
docker ps
```

---

 # Running the API

```
go run ./cmd/api
```

 The API will start on:

```
http://localhost:8080
```

 Health check:

```
GET /health
```

 Expected response:

```
{
  "message": "ok"
}
```

---

 # Running the Worker

 The worker will run as a separate process from the control plane.

 Example:

```
go run ./cmd/worker
```

 The worker connects to RabbitMQ and waits for deployment jobs.

```
RabbitMQ
    │
    ▼
Worker
    │
    ├── Git
    ├── Docker
    └── PostgreSQL
```

---

 # Example Deployment

 A simplified deployment might look like:

```
Project
│
├── Name: my-app
├── Repository: github.com/user/my-app
│
└── Deployments
    │
    ├── Deployment #1
    │   └── FAILED
    │
    ├── Deployment #2
    │   └── STOPPED
    │
    └── Deployment #3
        └── RUNNING
```

 Each deployment represents a specific attempt to build and run the project.

 This means deployments should generally be treated as historical records rather than simply overwriting the project's current state.

---

 # Design Principles

 The project follows a few important principles.

 ### Keep handlers thin

 Handlers deal with HTTP.

```
Handler
   ↓
Service
```

 Business logic should not live inside the HTTP handler.

 ### Keep repositories focused on persistence

 Repositories interact with PostgreSQL.

```
Service
   ↓
Repository
   ↓
Database
```

 ### Keep deployment state transitions in the domain

 Instead of changing status directly everywhere:

```
deployment.Status = models.RUNNING
```

 use:

```
err := deployment.TransitionTo(models.RUNNING)
```

 This prevents invalid state transitions from being scattered throughout the application.

 ### Use asynchronous workers for expensive operations

 Building Docker images and running deployments should not block an HTTP request.

 Instead:

```
API
 ↓
Queue
 ↓
Worker
```

 This makes the system easier to scale and more resilient.

---

 # Roadmap

 ## Phase 1 — API and Database

 - [x] User model
- [x] Project model
- [x] Deployment model
- [x] Project → Deployment relationship
- [x] Project CRUD
- [x] Deployment CRUD basics
- [x] Deployment state machine
- [x] Authentication middleware
- [ ] Authorization / resource ownership

 ## Phase 2 — Deployment Queue

 - [ ] RabbitMQ integration
- [ ] Deployment job model
- [ ] Publish deployment jobs
- [ ] Worker service
- [ ] Worker status updates
- [ ] Retry failed jobs

 ## Phase 3 — Docker

 - [ ] Clone Git repository
- [ ] Build Docker image
- [ ] Start application container
- [ ] Stop application container
- [ ] Remove old containers
- [ ] Dynamic port allocation
- [ ] Container health checks

 ## Phase 4 — GitHub Integration

 - [ ] GitHub repository connection
- [ ] Webhook endpoint
- [ ] Verify webhook signatures
- [ ] Create deployment on push
- [ ] Track commit SHA
- [ ] Automatic deployments

 ## Phase 5 — Reverse Proxy

 - [ ] Host-based routing
- [ ] Dynamic deployment discovery
- [ ] Route `*.localhost`
- [ ] Container health checking
- [ ] Remove routes for stopped deployments

 Example:

```
app1.localhost → :3001
app2.localhost → :3002
app3.localhost → :3003
```

 ## Phase 6 — Observability

 - [ ] Deployment logs
- [ ] Build logs
- [ ] Container logs
- [ ] Deployment history
- [ ] Build duration
- [ ] Resource usage
- [ ] Better error reporting

 ## Phase 7 — Production Concepts

 - [ ] Worker concurrency
- [ ] Job retries
- [ ] Dead-letter queue
- [ ] Idempotent deployments
- [ ] Distributed locking
- [ ] Resource limits
- [ ] CPU/memory limits
- [ ] Network isolation
- [ ] TLS
- [ ] Custom domains
- [ ] Secrets/environment variables

---

 # Future Architecture

 The long-term architecture is intended to evolve toward:

```
                         ┌───────────────┐
                         │    GitHub     │
                         └───────┬───────┘
                                 │
                              Webhook
                                 │
                                 ▼
                    ┌────────────────────────┐
                    │      Control Plane     │
                    │                        │
                    │ API                    │
                    │ Authentication         │
                    │ Projects               │
                    │ Deployments            │
                    │ Build Management       │
                    └───────┬─────────┬──────┘
                            │         │
                            │         ▼
                            │    PostgreSQL
                            │
                            ▼
                        RabbitMQ
                            │
                  ┌─────────┴─────────┐
                  │                   │
                  ▼                   ▼
              Worker 1             Worker 2
                  │                   │
                  └─────────┬─────────┘
                            │
                         Docker
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
           App #1         App #2        App #3
              │             │             │
              └─────────────┼─────────────┘
                            │
                            ▼
                      Reverse Proxy
                            │
                  ┌─────────┼─────────┐
                  ▼         ▼         ▼
             app1.local  app2.local  app3.local
```

---

 # Why Build This?

 This project is primarily a learning project focused on understanding how modern deployment platforms work.

 It combines several areas of backend and infrastructure engineering:

```
Go
 │
 ├── HTTP APIs
 ├── Authentication
 ├── Service architecture
 └── Concurrency

PostgreSQL
 │
 ├── Data modeling
 ├── Relationships
 └── Persistence

RabbitMQ
 │
 ├── Queues
 ├── Background jobs
 └── Asynchronous processing

Docker
 │
 ├── Images
 ├── Containers
 └── Application isolation

GitHub
 │
 └── Webhooks

Reverse Proxy
 │
 └── Dynamic routing
```

 The end goal is a small but understandable PaaS where:

```
git push
   ↓
GitHub
   ↓
Webhook
   ↓
Control Plane
   ↓
RabbitMQ
   ↓
Worker
   ↓
Docker
   ↓
Reverse Proxy
   ↓
my-app.localhost
```

---

 # License

 This project is currently a learning/experimental project.

 License information will be added as the project matures.