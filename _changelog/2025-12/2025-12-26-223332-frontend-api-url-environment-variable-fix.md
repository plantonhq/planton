# Frontend API URL Environment Variable Configuration Fix

**Date**: December 26, 2025
**Type**: Bug Fix / Configuration Enhancement
**Components**: Web App Frontend, Docker Configuration, Environment Variables

## Summary

Fixed environment variable mismatch between the frontend code and Docker configuration for the backend API URL, and made it configurable in the unified container deployment. The frontend was looking for `API_ENDPOINT` while Docker was setting `NEXT_PUBLIC_API_URL`, causing potential misconfiguration. Updated the code to use the correct Next.js convention (`NEXT_PUBLIC_API_URL`) and made it externally configurable like other environment variables.

## Problem Statement

The Planton web app's frontend code and Docker configuration had an inconsistency in environment variable naming:

### The Mismatch

**Frontend Code** (`app/frontend/src/app/layout.tsx`):
```typescript
const connectHost = process.env.API_ENDPOINT || 'http://localhost:50051';
```

**Docker Configuration** (`docker-compose.yml` and `Dockerfile.unified`):
```yaml
- NEXT_PUBLIC_API_URL=http://localhost:50051
```

**Issue**: The frontend was reading `API_ENDPOINT`, but Docker was setting `NEXT_PUBLIC_API_URL`. This meant:
- The environment variable set by Docker was being ignored
- The frontend always fell back to the hardcoded default
- No way to externally configure the backend API URL
- Not following Next.js conventions for public environment variables

### Why This Mattered

While the default value (`http://localhost:50051`) happened to work correctly for the unified container where both services run together, this created several issues:
1. **Silent failure**: The configuration appeared to work but was actually using fallback defaults
2. **Not configurable**: Users couldn't point the frontend to a different backend URL
3. **Inconsistent with Next.js best practices**: `NEXT_PUBLIC_*` prefix is the standard for client-accessible variables
4. **Inconsistent with other configs**: Other variables like `MONGODB_URI` were properly configurable

## Solution

Applied a two-part fix to standardize on `NEXT_PUBLIC_API_URL` and make it externally configurable:

### 1. Updated Frontend Code

Changed the environment variable reference in the root layout to use Next.js conventions:

**File**: `app/frontend/src/app/layout.tsx`
```typescript
// Before
const connectHost = process.env.API_ENDPOINT || 'http://localhost:50051';

// After
const connectHost = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:50051';
```

### 2. Made Docker Configuration Externally Configurable

Updated `docker-compose.yml` to allow external configuration with sensible defaults:

**File**: `docker-compose.yml`
```yaml
# Before (hardcoded)
- NEXT_PUBLIC_API_URL=http://localhost:50051

# After (configurable with default)
- NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:50051}
```

This follows the same pattern used for other configurable variables like:
- `MONGODB_URI=${MONGODB_URI:-mongodb://localhost:27017/planton}`
- `PULUMI_CONFIG_PASSPHRASE=${PULUMI_CONFIG_PASSPHRASE:-planton-default-passphrase}`

## Implementation Details

### Next.js Environment Variable Conventions

Next.js has special handling for variables prefixed with `NEXT_PUBLIC_*`:
- They are embedded at **build time** into the client bundle
- They are accessible in **browser-side** React components
- They must be set before running `next build`

This makes `NEXT_PUBLIC_API_URL` the correct choice for the backend API endpoint that the browser needs to connect to via gRPC-Web.

### Flow Through the Application

1. **Docker environment** sets `NEXT_PUBLIC_API_URL` (configurable or default)
2. **Next.js build** embeds the value into the production bundle
3. **RootLayout component** reads it from `process.env.NEXT_PUBLIC_API_URL`
4. **AppContextProvider** receives it as `connectHost` prop
5. **useConnectRpcClient hook** uses it to create the gRPC-Web transport
6. **All API calls** go to the configured backend URL

## Benefits

✅ **Consistency**: Frontend code now matches Docker configuration variable names

✅ **Configurability**: Users can override the backend URL via environment variables:
```bash
NEXT_PUBLIC_API_URL=https://api.custom.com docker-compose up
```

✅ **Next.js Best Practices**: Uses the correct `NEXT_PUBLIC_*` prefix for client-side variables

✅ **Pattern Alignment**: Follows the same configurable pattern as MongoDB and Pulumi settings

✅ **Flexibility**: Enables scenarios like:
- Running frontend separately from backend
- Pointing to remote backend APIs
- Testing with different backend environments
- Multi-environment deployments

## Usage Examples

### Default Unified Container
```bash
# Uses default http://localhost:50051
docker-compose up
```

### Custom Backend URL
```bash
# Via environment variable
NEXT_PUBLIC_API_URL=https://api.example.com docker-compose up

# Or via .env file
echo "NEXT_PUBLIC_API_URL=https://api.example.com" >> .env
docker-compose up
```

### Development with Separate Services
```bash
# Backend on different port
NEXT_PUBLIC_API_URL=http://localhost:8080 docker-compose up
```

## Impact

**Users**: Can now configure the backend API URL when deploying the web app, enabling flexible deployment architectures beyond the default unified container.

**Developers**: Clearer code that follows Next.js conventions and matches the actual Docker configuration.

**Operations**: Consistent configuration pattern across all environment variables in the unified container deployment.

## Files Changed

- `app/frontend/src/app/layout.tsx` - Updated environment variable reference
- `docker-compose.yml` - Made API URL configurable with default value

## Related Work

This fix complements the existing unified container deployment architecture documented in:
- `app/README.md` - Web app architecture and deployment modes
- `app/Dockerfile.unified` - Multi-stage build with MongoDB, backend, and frontend
- `docker-compose.yml` - Container orchestration configuration

---

**Status**: ✅ Production Ready
**Impact**: Low-risk fix with immediate benefits for configuration flexibility

