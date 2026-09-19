# Rankings API Documentation

## Overview

The Rankings API provides endpoints to retrieve boxer rankings based on various criteria. The system supports five ranking criteria:

1. **win_rate** - Win percentage (wins / total fights)
2. **total_fights** - Total career fights (wins + losses + draws)
3. **level** - Boxer level
4. **strength** - Strength stat
5. **power_score** - Composite score: `(wins * 2) + knockouts + (level * 3)`

## Endpoints

### Get Rankings

Retrieves a list of boxers ranked by the specified criteria.

```http
GET /rankings?criteria=win_rate&limit=100
```

**Query Parameters:**

| Parameter | Type   | Default   | Description                                           |
|-----------|--------|-----------|-------------------------------------------------------|
| criteria  | string | `win_rate`| Ranking criteria (see list above)                     |
| limit     | int    | `100`     | Number of results to return (max: 1000)              |

**Response:**

```json
{
  "criteria": "win_rate",
  "generated_at": "2026-09-19T12:00:00Z",
  "total_count": 50,
  "boxers": [
    {
      "id": 1,
      "name": "Mike Tyson",
      "nickname": "Iron Mike",
      "rank": 1,
      "wins": 50,
      "losses": 5,
      "draws": 2,
      "knockouts": 44,
      "level": 25,
      "strength": 95.5,
      "ranking_score": 0.877
    }
  ]
}
```

### Get Boxer Ranking

Retrieves a specific boxer's ranking position.

```http
GET /rankings/{boxer_id}?criteria=win_rate
```

**Path Parameters:**

| Parameter | Type | Description           |
|-----------|------|-----------------------|
| boxer_id  | int  | The boxer's ID        |

**Query Parameters:**

| Parameter | Type   | Default   | Description                           |
|-----------|--------|-----------|---------------------------------------|
| criteria  | string | `win_rate`| Ranking criteria to use for ranking   |

**Response:**

```json
{
  "rank": 5,
  "total_boxers": 120,
  "ranked_boxer": {
    "id": 42,
    "name": "Rocky Balboa",
    "nickname": "The Italian Stallion",
    "rank": 5,
    "wins": 35,
    "losses": 10,
    "draws": 3,
    "knockouts": 28,
    "level": 18,
    "strength": 82.0,
    "ranking_score": 0.75
  }
}
```

### Get Nearby Rankings

Retrieves boxers ranked near a specific boxer (useful for showing context).

```http
GET /rankings/nearby/{boxer_id}?criteria=win_rate&radius=5
```

**Path Parameters:**

| Parameter | Type | Description           |
|-----------|------|-----------------------|
| boxer_id  | int  | The boxer's ID        |

**Query Parameters:**

| Parameter | Type  | Default | Description                                        |
|-----------|-------|---------|----------------------------------------------------|
| criteria  | string| `win_rate`| Ranking criteria to use for ranking               |
| radius    | int   | `5`     | Number of boxers above and below to include        |

**Response:**

```json
{
  "criteria": "win_rate",
  "generated_at": "2026-09-19T12:00:00Z",
  "total_count": 11,
  "boxers": [
    {
      "id": 40,
      "name": "Boxer Above 1",
      "rank": 3,
      "wins": 45,
      "losses": 8,
      "draws": 1,
      "knockouts": 40,
      "level": 22,
      "strength": 90.0,
      "ranking_score": 0.849
    },
    {
      "id": 42,
      "name": "Rocky Balboa",
      "rank": 5,
      "wins": 35,
      "losses": 10,
      "draws": 3,
      "knockouts": 28,
      "level": 18,
      "strength": 82.0,
      "ranking_score": 0.75
    },
    {
      "id": 45,
      "name": "Boxer Below 1",
      "rank": 7,
      "wins": 30,
      "losses": 12,
      "draws": 2,
      "knockouts": 25,
      "level": 16,
      "strength": 80.0,
      "ranking_score": 0.714
    }
  ]
}
```

### Invalidate Rankings Cache (Admin)

Invalidates all cached rankings data. Useful after fights or significant boxer stat changes.

```http
POST /rankings/invalidate
```

**Authentication:** Required (JWT token)

**Response:**

```json
{
  "message": "Rankings cache invalidated successfully"
}
```

## Ranking Criteria Details

### Win Rate (`win_rate`)

Calculates win percentage as: `wins / (wins + losses)`

- Boxers with no fights get a score of 0
- Ties are broken by total wins (descending)

### Total Fights (`total_fights`)

Sums all career bouts: `wins + losses + draws`

- Measures activity and experience
- Ties are broken by wins (descending)

### Level (`level`)

Uses the boxer's current level field.

- Simple stat-based ranking
- Ties are broken by wins (descending)

### Strength (`strength`)

Uses the boxer's strength stat.

- Direct stat comparison
- Ties are broken by wins (descending)

### Power Score (`power_score`)

Composite metric: `(wins * 2) + knockouts + (level * 3)`

- Rewards winning, knocking out opponents, and progression
- Provides a holistic view of boxer performance

## Caching

Rankings are cached in Redis with a default TTL of 5 minutes. Cache keys follow these patterns:

- Rankings list: `rankings:{criteria}:limit:{limit}`
- Boxer ranking: `ranking:boxer:{boxer_id}:{criteria}`
- Nearby rankings: `rankings:nearby:boxer:{boxer_id}:{criteria}:radius:{radius}`

### Cache Invalidation

Cache should be invalidated when:

1. A fight completes (affects wins/losses/knockouts)
2. A boxer's stats are updated (strength, level)
3. A boxer is created or deleted

The `InvalidateRankingsCache()` method in `RankingsService` can be called to clear all cache entries. This should be integrated into the fight completion flow and boxer update operations.

## Performance Considerations

- All queries use LIMIT clauses to prevent excessive data loading
- Maximum limit is enforced at 1000 results
- Indexes on `wins`, `losses`, `draws`, `level`, `strength` columns improve query performance
- Redis caching significantly reduces database load for frequent ranking requests

## Error Responses

| Status Code | Description                    |
|-------------|--------------------------------|
| 400         | Invalid boxer ID or criteria   |
| 401         | Authentication required        |
| 404         | Boxer not found                |
| 500         | Internal server error          |

## Future Enhancements

- [ ] Add cache invalidation on fight completion
- [ ] Add cache invalidation on boxer stat updates
- [ ] Implement materialized views for complex rankings
- [ ] Add pagination support for large result sets
- [ ] Add filtering by level range or other criteria
- [ ] Add historical rankings tracking
