# notifications module

## Purpose
Stores and serves user notifications for matchmaking events.

## Technical Reason
Event records in the database provide durable, queryable notifications for likes and mutual matches, independent of transport delivery.

## Supported Events
- `liked`: when someone likes your profile
- `match`: when a mutual match is formed

## Endpoints
- `GET /notifications/unread?limit=50` (auth required)
- `POST /notifications/{notificationID}/read` (auth required)
