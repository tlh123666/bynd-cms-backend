# BYND CMS Backend

Independent Go API for the BYND administration console. It authenticates CMS operators and acts as a trusted gateway to the main BYND backend. The main backend's `CMS_INTERNAL_TOKEN` never reaches browser JavaScript.

## Included endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Service health check |
| `POST` | `/api/auth/login` | CMS administrator login |
| `GET` | `/api/user/info` | Current CMS operator |
| `GET` | `/api/community/groups` | Search and monitor groups |
| `GET` | `/api/community/groups/:groupId` | Group summary |
| `GET` | `/api/community/challenges` | List all challenge states |
| `GET` | `/api/community/challenges/:challengeId` | Challenge and daily content |
| `POST` | `/api/community/challenges` | Create a challenge draft |
| `PATCH` | `/api/community/challenges/:challengeId` | Edit challenge metadata |
| `PUT` | `/api/community/challenges/:challengeId/days/:dayNumber` | Save daily content |
| `POST` | `/api/community/challenges/:challengeId/publish` | Publish a complete challenge |
| `POST` | `/api/community/challenges/:challengeId/archive` | Archive a challenge |

All endpoints except login and health require the CMS bearer token.

## Run locally

1. Copy `.env.example` to `.env`.
2. Set `CMS_INTERNAL_TOKEN` to the same value used by `BYND-backend`.
3. Replace the development JWT secret and administrator password.
4. Start the API:

```powershell
go mod tidy
go run ./cmd/server
```

The default address is `http://localhost:8082`. The Vue CMS development proxy already targets this address.

## Production

- Use strong secret values supplied by the deployment environment.
- Prefer `CMS_ADMIN_PASSWORD_HASH` (bcrypt) over a plaintext production password.
- Put this service behind HTTPS and proxy `/api` from the CMS domain to it.
- Restrict `CMS_ALLOWED_ORIGINS` to the exact CMS origins.
- Rotate `CMS_INTERNAL_TOKEN` and `CMS_JWT_SECRET` independently.
