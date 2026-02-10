# profiles module

## Purpose
Manages user profile CRUD data used by matchmaking and profile browsing.

## Technical Reason
Separating profile concerns keeps profile validation/privacy logic independent from auth and matchmaking ranking flows.

## Fields Managed
- age
- profession
- education
- income
- residency_status (visa/residency)
- location
- marital_status
- description

## Privacy Flags
- hide_contact_info
- hide_address
- hide_income
- hide_visa_status

## Endpoints
- `POST /profiles/` (auth required): create own profile
- `PUT /profiles/{userID}` (auth required): update own profile
- `GET /profiles/{userID}`: get profile by user id
- `GET /profiles/?page=1&size=20`: list paginated profiles

Only the authenticated user can create/update their own `user_id`.

## Example JSON

### Create Profile Request
```json
{
  "user_id": "u1",
  "age": 29,
  "profession": "Software Engineer",
  "education": "Masters",
  "income": 120000,
  "residency_status": "Citizen",
  "location": "Berlin",
  "marital_status": "Single",
  "description": "I enjoy travel and books.",
  "hide_contact_info": true,
  "hide_address": false,
  "hide_income": true,
  "hide_visa_status": false
}
```

### Create/Update/Get Response
```json
{
  "user_id": "u1",
  "age": 29,
  "profession": "Software Engineer",
  "education": "Masters",
  "income": 120000,
  "residency_status": "Citizen",
  "location": "Berlin",
  "marital_status": "Single",
  "description": "I enjoy travel and books.",
  "hide_contact_info": true,
  "hide_address": false,
  "hide_income": true,
  "hide_visa_status": false,
  "created_at": "2026-02-09T10:00:00Z",
  "updated_at": "2026-02-09T10:00:00Z"
}
```

### List Response
```json
{
  "items": [
    {
      "user_id": "u1",
      "age": 29,
      "profession": "Software Engineer",
      "education": "Masters",
      "income": 120000,
      "residency_status": "Citizen",
      "location": "Berlin",
      "marital_status": "Single",
      "description": "I enjoy travel and books.",
      "hide_contact_info": true,
      "hide_address": false,
      "hide_income": true,
      "hide_visa_status": false,
      "created_at": "2026-02-09T10:00:00Z",
      "updated_at": "2026-02-09T10:00:00Z"
    }
  ],
  "page": 1,
  "size": 20
}
```
