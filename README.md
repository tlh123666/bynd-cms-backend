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


## Wellness schema updates

The CMS gateway exposes read-only data from the main BYND database; it does not own or migrate a second copy of health tables.

| Method | CMS endpoint | Data |
| --- | --- | --- |
| GET | `/api/wellness/vitality` | Paginated persisted recovery/energy observations and filtered statistics |
| GET | `/api/wellness/vitality/:userId` | Active user identity and currently saved preferences; missing overrides remain null |
| GET | `/api/wellness/heart-measurements` | Paginated PPG/ECG measurement metadata, including records without a daily summary |

Existing sleep, heart-rate and diary detail endpoints now include saved preferences/settings, measurement metadata, workspace notes/revision, supported event details and tracker overrides. No health reading, score or preference is modified by these endpoints. Private photo keys and raw waveforms are excluded.

Vitality filters: `page`, `pageSize` (up to 100), `keyword`, `startDate`, `endDate`, `kind=recovery|energy`, `considerTraining=true|false`, `algorithmVersion`. Measurement filters: pagination/user/date filters plus `waveformKind=ppg|ecg|unknown` and `measurementContext=rest|after_coffee|after_workout`. Dates for these two timestamp-based lists are inclusive UTC calendar days. Daily summaries and diaries retain their stored local `statDate`.

Requests forward `Accept-Language`; wellness errors support Chinese, English and Portuguese. Upstream business failures cannot be silently returned as successful data, and internal authentication failures do not clear an operator's valid CMS session.

Deploy the main backend with migrations `000061` through `000065`, then apply `000066_cms_vitality_observation_index.up.sql` using its normal migration workflow. Restart both Go services and publish the updated CMS frontend together. The new migration adds an observation time index only; the gateway needs no additional environment variables.

See [CMS page and verification notes](../bynd-cms/docs/wellness-management.md). Run `go test ./...` here and `go test ./internal/module/cmshealth/...` in the main backend. The optional main-backend MySQL suite uses `BYND_TEST_MYSQL_DSN` and creates/removes an isolated test database.
