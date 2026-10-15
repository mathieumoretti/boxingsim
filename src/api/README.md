# Frontend API Layer (MAT-120)

This directory contains the frontend API client layer for communicating with the boxing simulation backend. The architecture follows a three-layer pattern to reduce duplication and provide consistent error handling.

## Architecture

```
Components → Hooks → API Functions → HTTP Client → Backend
```

### Layer 1: HTTP Client (`client.js`)
Base fetch wrapper with authentication, error handling, and configuration.

**Exports:**
- `apiRequest(url, options, config)` - Core request method
- `apiGetJson()`, `apiPostJson()`, `apiPutJson()`, `apiDelete()` - Convenience methods
- `ApiError` - Structured error class
- `getToken()`, `setToken()`, `clearToken()`, `getUser()` - Auth utilities

### Layer 2: Domain APIs (`*.js`)
Endpoint-specific functions for each domain.

**Modules:**
- `fights.js` - Fight operations (history, details, booking)
- `rankings.js` - Rankings with pagination
- `opponents.js` - Opponent discovery and validation
- `training.js` - Training sessions
- `world.js` - Game world time
- `users.js` - Auth and user operations

### Layer 3: React Hooks (`../hooks/`)
State management for queries (separate directory).

**Hooks:**
- `useBoxer(boxerId)` - Fetch boxer data
- `useRankings(options)` - Rankings with pagination

## Usage Examples

### Basic API Call (Layer 2)

```js
import { fetchFightHistory } from '../api/fights';

async function loadFights() {
  try {
    const fights = await fetchFightHistory(123);
    console.log(fights);
  } catch (error) {
    if (error instanceof ApiError) {
      console.error(`HTTP ${error.status}:`, error.body);
    }
  }
}
```

### With React Hook (Layer 3)

```js
import { useRankings } from '../hooks/useRankings';

function RankingsPage() {
  const { data, isLoading, error, pagination, setPage } = useRankings({ 
    criteria: 'win_rate',
    page: 1 
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div>
      {data.map(boxer => <BoxerCard key={boxer.id} boxer={boxer} />)}
      {/* Pagination controls */}
    </div>
  );
}
```

### Direct HTTP Client (Layer 1)

```js
import { apiGetJson, apiPostJson, ApiError } from '../api/client';

// GET request
const boxer = await apiGetJson('/api/boxers/123');

// POST with body
const result = await apiPostJson('/api/boxers/123/train', {
  training_type_id: 1
});

// With custom config (retry logic)
const data = await apiRequest('/api/unreliable-endpoint', {
  method: 'GET'
}, {
  maxRetries: 3,
  retryDelayMs: 1000
});
```

## Error Handling

### ApiError Class

All HTTP errors throw `ApiError` with structured data:

```js
import { ApiError } from '../api/client';

try {
  await apiGetJson('/api/endpoint');
} catch (error) {
  if (error instanceof ApiError) {
    console.log(error.status);     // HTTP status code (404, 500, etc.)
    console.log(error.statusText); // "Not Found", "Internal Server Error"
    console.log(error.body);       // Parsed JSON body or { message: string }
    console.log(error.message);    // User-friendly message
  }
}
```

### Automatic 401 Handling

When the server returns 401 Unauthorized:
1. Token is cleared from localStorage
2. User is redirected to `/login`
3. `ApiError` is thrown (catch if you need to handle it)

To disable auto-redirect:

```js
const response = await apiRequest('/api/silent-check', {}, {
  redirectOnUnauth: false
});
// Handle 401 manually...
```

## Configuration Options

All `api*` functions accept a third `config` parameter:

| Option | Default | Description |
|--------|---------|-------------|
| `throwOnHttpError` | `true` | Throw `ApiError` on non-2xx responses |
| `maxRetries` | `0` | Number of retry attempts on network errors |
| `retryDelayMs` | `1000` | Initial delay between retries (exponential backoff) |
| `redirectOnUnauth` | `true` | Auto-redirect to login on 401 |

Example:

```js
const data = await apiGetJson('/api/endpoint', {}, {
  maxRetries: 3,
  retryDelayMs: 2000
});
```

## API Endpoints Reference

### Fights (`fights.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `fetchFightHistory(boxerId)` | `GET /api/boxers/:id/fights-history` | Get boxer's fight history |
| `fetchUpcomingFight(boxerId)` | `GET /api/boxers/:id/upcoming-fight` | Get next scheduled fight |
| `fetchActiveFights(boxerIds)` | `GET /api/fights/active` | Get all active fights (filtered) |
| `fetchFightDetails(fightId)` | `GET /api/fights/:id` | Get fight details with play-by-play |
| `fetchBoxer(boxerId)` | `GET /api/boxers/:id` | Get boxer details |
| `bookFight(boxerId, data)` | `POST /api/boxers/:id/fights` | Schedule a new fight |

### Rankings (`rankings.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `fetchRankings(options)` | `GET /api/rankings` | Get paginated rankings |

Options: `{ criteria, page, limit }`

### Opponents (`opponents.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `fetchOpponents(boxerId, filters)` | `GET /api/boxers/:id/opponents` | Get opponent list |
| `fetchBestMatch(boxerId)` | `GET /api/opponents/best_match` | Get best opponent match |
| `validateFight(data)` | `GET /api/opponents/validate` | Validate fight can be booked |

### Training (`training.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `fetchTrainingTypes()` | `GET /api/training-types` | Get all training types |
| `fetchActiveTrainingSessions(boxerId)` | `GET /api/boxers/:id/training-sessions` | Get active sessions |
| `startTrainingSession(boxerId, data)` | `POST /api/boxers/:id/train` | Start new training |
| `completeActiveTrainingSessions(boxerId)` | `POST /api/training/bulk-complete` | Complete all training |

### World (`world.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `fetchWorldTime()` | `GET /api/world/time` | Get current game time |
| `getCurrentGameTime()` | `GET /api/world/time` | Convenience: get time string only |

### Users/Auth (`users.js`)

| Function | Endpoint | Description |
|----------|----------|-------------|
| `registerUser(data)` | `POST /api/auth/register` | Register new user |
| `login(credentials)` | `POST /api/auth/login` | User login |
| `getCurrentUser()` | `GET /api/auth/me` | Get current user |
| `fetchUserBoxers(userId)` | `GET /api/users/:id/boxers` | Get user's boxers |

## Migration from Old Pattern

**Before (`src/utils/`):**

```js
import { authenticatedFetch } from '../utils/auth';

const response = await authenticatedFetch('/api/boxers/123');
if (!response.ok) throw new Error('Failed');
const data = await response.json();
```

**After (API layer):**

```js
import { fetchBoxer } from '../api/fights';

const boxer = await fetchBoxer(123); // Throws ApiError on failure
```

## Testing

### Mock API Calls

For component tests, mock the API functions:

```js
import { renderHook } from '@testing-library/react';
import { useBoxer } from '../hooks/useBoxer';

// Mock the API function
jest.mock('../api/fights', () => ({
  fetchBoxer: jest.fn(() => ({ id: 1, name: 'Test Boxer' })),
  fetchUpcomingFight: jest.fn(() => null)
}));

const { result } = renderHook(() => useBoxer(123));
expect(result.current.boxer.name).toBe('Test Boxer');
```

## Backward Compatibility

Old `src/utils/auth.js` and `src/utils/fights.js` files are preserved for backward compatibility during gradual migration. New code should use the `src/api/` modules.

## Related

- [MAT-120 Issue](https://linear.app/mathieu-moretti/issue/MAT-120/frontend-apidata-access-layer)
- [[mat-120-audit-findings]] - Initial audit findings
- [[mat-120-api-layer-complete]] - Implementation documentation
