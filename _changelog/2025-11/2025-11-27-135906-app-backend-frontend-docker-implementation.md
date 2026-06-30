# App Backend, Frontend, and Docker Implementation

**Date**: November 27, 2025  
**Type**: Feature Addition  
**Components**: Backend Service, Frontend Web Application, Docker Containerization  
**Impact**: New Application Infrastructure

---

## Summary

Implemented a complete full-stack application in the `/app` directory with a Go-based backend service using Connect RPC and MongoDB, a Next.js frontend with Material-UI, and production-ready Docker containerization for both services. The application provides a deployment component management system with a modern web interface.

---

## Problem Statement / Motivation

Planton needed a web application to manage and visualize deployment components. The requirements included:

- **Backend API**: A gRPC-compatible service to manage deployment component data
- **Frontend Interface**: A modern web UI for viewing and managing deployment components
- **Containerization**: Production-ready Docker images for easy deployment
- **Integration**: Seamless communication between frontend and backend using Connect RPC

---

## Solution / What's New

### 1. Backend Service (`app/backend/`)

A Go-based backend service providing deployment component management via Connect RPC.

#### Core Components

**Server Implementation** (`cmd/server/main.go`):

- HTTP/2 server with h2c (HTTP/2 Cleartext) support for gRPC-Web compatibility
- Graceful shutdown with 5-second timeout
- CORS middleware configured for cross-origin requests
- Environment-based configuration (port, MongoDB connection)
- Structured logging with logrus

**Service Layer** (`internal/service/deployment_component_service.go`):

- `DeploymentComponentService` implementing Connect RPC interface
- `ListDeploymentComponents` RPC method with optional filtering by provider and kind
- Request/response transformation between MongoDB models and protobuf messages
- Timestamp conversion using `google.protobuf.Timestamp`

**Database Layer** (`internal/database/`):

- MongoDB connection management with connection pooling
- `DeploymentComponentRepository` for data access operations
- Filtering support for provider and kind fields
- Error handling and connection lifecycle management

**Data Models** (`pkg/models/deployment_component.go`):

- `DeploymentComponent` struct with BSON tags for MongoDB
- Fields: ID, Kind, Provider, Name, Version, IDPrefix, IsServiceKind, CreatedAt, UpdatedAt
- Primitive ObjectID for MongoDB document identification

**Protocol Buffers** (`apis/proto/deployment_component_service.proto`):

- `DeploymentComponentService` service definition
- `ListDeploymentComponents` RPC with optional filters
- `DeploymentComponent` message with all metadata fields
- Generated Connect RPC handlers and protobuf code

**Build System** (`Makefile`):

- `make deps` - Download Go dependencies
- `make generate` - Generate protobuf code via buf
- `make build` - Build server binary
- `make run` - Run built server
- `make dev` - Development mode with hot reload
- `make clean` - Remove build artifacts
- `make test` - Run test suite

#### Technology Stack

- **Language**: Go 1.24.7
- **RPC Framework**: Connect RPC (gRPC-compatible)
- **Database**: MongoDB (via mongo-driver)
- **Logging**: logrus
- **HTTP/2**: golang.org/x/net/http2
- **CORS**: github.com/rs/cors

### 2. Frontend Application (`app/frontend/`)

A Next.js 14 web application with Material-UI providing a modern dashboard interface.

#### Core Components

**Dashboard Page** (`src/app/dashboard/page.tsx`):

- Deployment component data table with sorting, pagination, and filtering
- Statistics cards showing total products, inventory, and average price
- Real-time data loading from backend API
- Action buttons (View, Edit, Delete) for each component
- Refresh functionality with loading states
- Error handling with user-friendly alerts

**Layout System** (`src/components/layout/`):

- Header component with navigation
- Sidebar with collapsible menu
- Responsive layout using Material-UI Grid2
- Theme-aware styling with Emotion

**Data Table Component** (`src/components/shared/data-table/`):

- Reusable data table with column definitions
- Sortable columns
- Row selection (single and multi-select)
- Pagination controls
- Custom cell rendering support
- Action buttons per row

**Connect RPC Integration** (`src/hooks/useConnectRpcClient.ts`):

- Custom React hook for Connect RPC client creation
- Automatic client initialization based on service and host
- Global error handling with message sanitization
- Binary format support for efficient data transfer
- Context-based host configuration

**Query Services** (`src/app/dashboard/_services/query.ts`):

- `useDashboardQuery` hook for API interactions
- `listDeploymentComponents` query method
- Loading state management via AppContext
- Snackbar notifications for errors
- Promise-based API with proper error handling

**Theme System** (`src/themes/`):

- Light and dark theme support
- Color palette definitions
- Material-UI theme configuration
- Cookie-based theme persistence
- Server-side rendering support for theme

**App Context** (`src/contexts/appContext.tsx`):

- Global state management
- Connect RPC host configuration
- Page loading state
- Snackbar notifications
- Theme mode management
- Navbar state persistence

**Health Check Endpoint** (`pages/api/health.js`):

- REST endpoint for Docker health checks
- Returns service status and timestamp
- Used by container orchestration systems

**Next.js Configuration** (`next.config.js`):

- Standalone output mode for Docker optimization
- Emotion compiler configuration for CSS-in-JS
- Webpack cache disabled for development
- Environment variable support via dotenv

#### Technology Stack

- **Framework**: Next.js 14.2.14 (App Router)
- **UI Library**: Material-UI (MUI) 6.1.2
- **Styling**: Emotion (CSS-in-JS)
- **RPC Client**: @connectrpc/connect-web 2.0.2
- **Protobuf**: @bufbuild/protobuf 2.5.1
- **Language**: TypeScript 5.6.2
- **Package Manager**: Yarn 3.6.4

#### Build System (`Makefile`)

- `make deps` - Install dependencies
- `make generate` - Generate protobuf stubs
- `make build` - Build production bundle
- `make dev` - Start development server
- `make clean` - Remove build artifacts
- `make update-deps` - Update dependencies

### 3. Docker Containerization

Production-ready multi-stage Docker builds for both backend and frontend.

#### Backend Dockerfile (`app/backend/Dockerfile`)

**Builder Stage**:

- Base image: `golang:1.24.7-alpine`
- Installs build dependencies (git, make, curl)
- Installs buf tool (v1.45.0) for protobuf code generation
- Copies go.mod/go.sum for dependency caching
- Downloads Go modules
- Generates protobuf code
- Builds statically linked binary (CGO_ENABLED=0)

**Runtime Stage**:

- Base image: `alpine:3.19`
- Installs ca-certificates for HTTPS, tzdata for timezone, procps for healthcheck
- Creates non-root user (appuser:appgroup, UID 1001)
- Copies binary from builder stage
- Sets proper file ownership
- Exposes port 50051
- Health check using `pgrep` to verify process is running
- Runs as non-root user for security

**Features**:

- Multi-stage build for minimal image size
- Non-root user execution
- Health check support
- Statically linked binary (no runtime dependencies)
- Optimized layer caching

#### Frontend Dockerfile (`app/frontend/Dockerfile`)

**Dependencies Stage**:

- Base image: `node:20-alpine`
- Enables Corepack for Yarn 3
- Copies package.json and yarn.lock
- Installs dependencies with frozen lockfile

**Builder Stage**:

- Copies node_modules from deps stage
- Copies source code
- Generates protobuf files (if buf.gen.yaml exists)
- Builds Next.js application with standalone output
- Disables Next.js telemetry

**Runtime Stage**:

- Base image: `node:20-alpine`
- Installs wget for healthcheck
- Creates non-root user (nextjs:nodejs, UID 1001)
- Copies public assets
- Copies standalone build output
- Sets proper permissions for .next directory
- Exposes port 3000
- Health check using wget to verify HTTP endpoint
- Runs as non-root user

**Features**:

- Multi-stage build for minimal image size
- Standalone output mode (only necessary files)
- Non-root user execution
- HTTP health check endpoint
- Optimized layer caching
- Production environment variables

#### Docker Benefits

- **Security**: Both containers run as non-root users
- **Size**: Multi-stage builds minimize final image size
- **Health**: Built-in health checks for orchestration
- **Caching**: Optimized layer caching for faster rebuilds
- **Production-Ready**: Follows Docker best practices

---

## Implementation Details

### Backend Architecture

```
app/backend/
├── apis/
│   ├── proto/
│   │   └── deployment_component_service.proto
│   └── gen/
│       └── go/
│           └── proto/
│               ├── backendv1connect/
│               │   └── deployment_component_service.connect.go
│               └── deployment_component_service.pb.go
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── database/
│   │   ├── mongodb.go
│   │   └── deployment_component_repo.go
│   └── service/
│       └── deployment_component_service.go
├── pkg/
│   └── models/
│       └── deployment_component.go
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

### Frontend Architecture

```
app/frontend/
├── pages/
│   └── api/
│       └── health.js
├── public/
├── src/
│   ├── app/
│   │   ├── dashboard/
│   │   │   ├── _services/
│   │   │   │   └── query.ts
│   │   │   ├── page.tsx
│   │   │   └── styled.ts
│   │   ├── layout.tsx
│   │   └── page.tsx
│   ├── components/
│   │   ├── layout/
│   │   │   ├── header/
│   │   │   ├── sidebar/
│   │   │   └── layout.tsx
│   │   ├── providers/
│   │   │   └── index.tsx
│   │   └── shared/
│   │       └── data-table/
│   ├── contexts/
│   │   ├── appContext.tsx
│   │   ├── models.ts
│   │   └── index.ts
│   ├── gen/
│   │   └── proto/
│   │       ├── deployment_component_service_connect.ts
│   │       └── deployment_component_service_pb.ts
│   ├── hooks/
│   │   ├── useConnectRpcClient.ts
│   │   └── index.ts
│   ├── lib/
│   │   ├── cookie-constants.ts
│   │   ├── cookie-utils.ts
│   │   ├── server/
│   │   │   └── cookies.ts
│   │   └── utils.ts
│   └── themes/
│       ├── colors.ts
│       ├── dark-colors.ts
│       ├── dark.tsx
│       ├── light-colors.ts
│       ├── light.tsx
│       └── theme.ts
├── Dockerfile
├── Makefile
├── next.config.js
├── package.json
├── tsconfig.json
└── yarn.lock
```

### Protocol Buffer Integration

**Backend**:

- Proto definitions in `apis/proto/`
- Code generation via buf tool
- Generated Go stubs in `apis/gen/go/proto/`
- Connect RPC handlers auto-generated

**Frontend**:

- Shared proto definitions (via backend generation or shared repo)
- Generated TypeScript stubs in `src/gen/proto/`
- Type-safe RPC calls using generated schemas
- Binary format for efficient data transfer

### Environment Configuration

**Backend Environment Variables**:

- `MONGODB_URI` - MongoDB connection string (default: `mongodb://localhost:27017`)
- `MONGODB_DATABASE` - Database name (default: `planton`)
- `SERVER_PORT` or `PORT` - Server port (default: `50051`)

**Frontend Environment Variables**:

- `API_ENDPOINT` - Backend API URL (default: `http://localhost:50051`)
- `NODE_ENV` - Node environment (production/development)

### Data Flow

1. **Frontend Request**:

   - User interacts with dashboard
   - `useDashboardQuery` hook called
   - Connect RPC client created with transport
   - Request sent to backend via HTTP/2

2. **Backend Processing**:

   - Connect RPC handler receives request
   - Service layer processes request
   - Repository queries MongoDB with filters
   - Results transformed to protobuf messages
   - Response sent back to frontend

3. **Frontend Display**:
   - Response received and deserialized
   - State updated with deployment components
   - Data table renders with sorting/pagination
   - User sees updated information

---

## Benefits

### Developer Experience

- **Type Safety**: Full TypeScript/Go type safety across the stack
- **Code Generation**: Protobuf eliminates manual serialization code
- **Hot Reload**: Development servers support fast iteration
- **Clear Structure**: Well-organized codebase with separation of concerns

### Production Readiness

- **Containerization**: Docker images ready for deployment
- **Security**: Non-root users, minimal attack surface
- **Health Checks**: Built-in monitoring support
- **Scalability**: Stateless backend, stateless frontend
- **Performance**: HTTP/2, binary protobuf, optimized builds

### Maintainability

- **Modular Design**: Clear separation between layers
- **Standard Patterns**: Follows Go and React best practices
- **Documentation**: Code is self-documenting with clear structure
- **Testing**: Makefile targets for running tests

---

## Usage

### Backend Development

```bash
cd app/backend

# Install dependencies
make deps

# Generate protobuf code
make generate

# Run in development mode
make dev

# Build binary
make build

# Run built binary
make run
```

### Frontend Development

```bash
cd app/frontend

# Install dependencies
make deps

# Generate protobuf stubs
make generate

# Run development server
make dev

# Build for production
make build
```

### Docker Deployment

**Backend**:

```bash
cd app/backend
docker build -t planton-backend .
docker run -p 50051:50051 \
  -e MONGODB_URI=mongodb://mongodb:27017 \
  -e MONGODB_DATABASE=planton \
  planton-backend
```

**Frontend**:

```bash
cd app/frontend
docker build -t planton-frontend .
docker run -p 3000:3000 \
  -e API_ENDPOINT=http://backend:50051 \
  planton-frontend
```

### Docker Compose (Example)

```yaml
version: '3.8'
services:
  mongodb:
    image: mongo:7
    ports:
      - '27017:27017'
    volumes:
      - mongodb_data:/data/db

  backend:
    build: ./app/backend
    ports:
      - '50051:50051'
    environment:
      - MONGODB_URI=mongodb://mongodb:27017
      - MONGODB_DATABASE=planton
    depends_on:
      - mongodb

  frontend:
    build: ./app/frontend
    ports:
      - '3000:3000'
    environment:
      - API_ENDPOINT=http://backend:50051
    depends_on:
      - backend

volumes:
  mongodb_data:
```

---

## Technical Decisions

### Why Connect RPC?

- **gRPC Compatibility**: Works with gRPC services
- **Web Support**: Built-in gRPC-Web support for browsers
- **Type Safety**: Generated code ensures type safety
- **Performance**: Binary protocol, HTTP/2 multiplexing
- **Simplicity**: Easier than raw gRPC for web clients

### Why MongoDB?

- **Flexibility**: Schema-less design for evolving data models
- **Document Model**: Natural fit for deployment component metadata
- **Querying**: Rich query capabilities for filtering
- **Scalability**: Horizontal scaling support
- **Maturity**: Well-established Go driver

### Why Next.js?

- **React Framework**: Modern React with App Router
- **SSR/SSG**: Server-side rendering and static generation
- **Performance**: Built-in optimizations
- **Developer Experience**: Excellent tooling and hot reload
- **Standalone Mode**: Docker-friendly output

### Why Multi-Stage Docker Builds?

- **Size**: Final images contain only runtime dependencies
- **Security**: Minimal attack surface
- **Caching**: Better layer caching for faster rebuilds
- **Best Practice**: Industry standard for production images

---

## File Structure Summary

### Backend Files

- **Go Source**: ~500 lines across 6 files
- **Protobuf**: 1 service definition, 3 message types
- **Dockerfile**: 57 lines, multi-stage build
- **Makefile**: 42 lines, 7 targets
- **Dependencies**: 6 direct, 9 indirect

### Frontend Files

- **TypeScript/TSX**: ~2,000+ lines across 30+ files
- **Components**: 10+ reusable components
- **Hooks**: Custom React hooks for RPC
- **Dockerfile**: 77 lines, multi-stage build
- **Makefile**: 24 lines, 6 targets
- **Dependencies**: 20+ direct, 100+ indirect

---

## Future Enhancements

### Potential Additions

1. **Authentication**: User authentication and authorization
2. **CRUD Operations**: Create, update, delete deployment components
3. **Real-time Updates**: WebSocket support for live data
4. **Advanced Filtering**: More sophisticated query capabilities
5. **Export/Import**: Data export and import functionality
6. **Analytics**: Usage tracking and analytics dashboard
7. **Testing**: Comprehensive unit and integration tests
8. **CI/CD**: Automated build and deployment pipelines

---

## Related Work

This implementation establishes the foundation for Planton's web application infrastructure. It integrates with:

- **Planton CLI**: Deployment component definitions
- **Protobuf APIs**: Shared API definitions
- **Docker Ecosystem**: Container orchestration platforms

---

## Status

**Status**: ✅ Production Ready  
**Backend**: Fully functional with MongoDB integration  
**Frontend**: Complete dashboard with data visualization  
**Docker**: Production-ready containerization  
**Documentation**: Comprehensive code structure  
**Timeline**: Completed November 27, 2025

---
