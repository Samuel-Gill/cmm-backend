# matches module

## Purpose
Implements matchmaking browse/filter, likes, and mutual match relationship retrieval.

## Technical Reason
Separating matchmaking interactions into a dedicated module allows query optimization and relationship state logic to evolve without coupling to profile CRUD handlers.

## Endpoints
- `GET /matches/browse?min_age=25&max_age=35&gender=female&min_income=50000&max_income=200000&location=Berlin&residency_status=Citizen&page=1&size=20`
- `POST /matches/like`
- `GET /matches/relationships`

All endpoints require authentication.

## Example API Calls

### Browse Profiles
```bash
curl -H "Authorization: Bearer <access-token>" \
  "http://localhost:8080/matches/browse?min_age=25&max_age=35&gender=female&min_income=50000&max_income=200000&location=Berlin&residency_status=Citizen&page=1&size=20"
```

### Like Profile
```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer <access-token>" \
  -d '{"liked_user_id":"user-123"}' \
  http://localhost:8080/matches/like
```

### Relationship Summary
```bash
curl -H "Authorization: Bearer <access-token>" \
  http://localhost:8080/matches/relationships
```

### Sample Relationship Response
```json
{
  "liked_by": ["user-2", "user-7"],
  "liked": ["user-5", "user-7"],
  "mutual_matches": ["user-7"]
}
```
