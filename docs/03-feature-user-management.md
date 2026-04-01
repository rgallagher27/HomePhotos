# Feature: User Management

## Overview

Authentication and user management in HomePhotos is handled by [Clerk](https://clerk.com/), a managed authentication service. HomePhotos does not implement its own login system, password storage, or session management. Instead, Clerk handles all identity concerns, while HomePhotos stores only app-specific data (role, preferences) locally in SQLite, linked to Clerk via a `clerk_user_id` foreign key.

This separation keeps the security-critical auth layer out of the HomePhotos codebase and delegates it to a purpose-built service with ongoing security maintenance.

## Clerk Integration

### What Clerk Handles

- **User creation** -- new accounts are created and stored in Clerk's infrastructure
- **Login / signup UI** -- embedded Clerk components render the sign-in and sign-up forms directly in the SvelteKit frontend
- **Session management** -- Clerk tracks active sessions, handles token refresh, and provides session revocation
- **JWT issuance** -- Clerk issues signed JWTs that the backend verifies on every API request
- **Password reset** -- self-service password reset flow, fully managed by Clerk
- **Email verification** -- email verification during signup, managed by Clerk

### Frontend Integration

Use Clerk's SvelteKit SDK (or the JavaScript SDK) to embed sign-in and sign-up components into the SvelteKit app. Clerk provides pre-built UI components that handle the full authentication flow:

- `<SignIn />` -- renders the sign-in form
- `<SignUp />` -- renders the sign-up form
- `<UserButton />` -- renders the signed-in user's avatar with a dropdown for account management and sign-out

Clerk's SDK also provides reactive stores/hooks to access the current user and session state throughout the frontend.

### Backend Integration

The Go backend verifies Clerk session JWTs on every incoming API request. Two approaches are available:

1. **Clerk Go SDK** -- use Clerk's official Go SDK to verify tokens and extract claims
2. **JWKS endpoint** -- fetch Clerk's public keys from the JWKS endpoint and verify JWTs using a standard JWT library

The backend never needs to call Clerk's API for normal request authentication. JWT verification is done locally using cached public keys.

## Local User Model

HomePhotos maintains a local `users` table in SQLite to store app-specific data that Clerk does not manage. The `clerk_user_id` column links back to the Clerk-managed identity.

### SQLite Schema

```sql
CREATE TABLE users (
    id              INTEGER PRIMARY KEY,
    clerk_user_id   TEXT UNIQUE NOT NULL,
    role            TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'viewer')),
    display_name    TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login      DATETIME
);
```

| Column | Type | Description |
|---|---|---|
| `id` | `INTEGER PRIMARY KEY` | Local auto-incrementing ID |
| `clerk_user_id` | `TEXT UNIQUE NOT NULL` | Clerk's user ID (e.g., `user_2abc...`). Links to Clerk identity. |
| `role` | `TEXT NOT NULL` | Either `'admin'` or `'viewer'`. Defaults to `'viewer'`. |
| `display_name` | `TEXT` | Optional display name (may be synced from Clerk or set locally) |
| `created_at` | `DATETIME` | When the local record was created |
| `last_login` | `DATETIME` | Updated on each authenticated API request |

## Roles

### Admin

Admins have full control over the HomePhotos application:

- Manage tags and tag groups (create, rename, delete, organize)
- Trigger library rescans (manual scan of the source directory)
- Manage user roles (promote viewers to admin, demote admins to viewer)
- View system status (cache progress, scan status, error logs)

### Viewer

Viewers have read-only access to the photo library:

- Browse photos in timeline, folder, and tag views
- View photo detail with full-resolution image
- View photo EXIF metadata (camera, lens, settings, date)
- View tags assigned to photos

## Authentication Flow

The full authentication flow from initial visit to authorized API request:

1. **User visits HomePhotos** -- the SvelteKit app loads and the Clerk sign-in component renders if the user is not already authenticated.
2. **User authenticates via Clerk** -- the user enters their credentials (email/password, or another method configured in the Clerk dashboard).
3. **Clerk issues a session JWT** -- upon successful authentication, Clerk creates a session and issues a signed JWT containing claims such as the Clerk user ID and session expiry.
4. **Frontend includes JWT in API requests** -- the SvelteKit frontend attaches the JWT to every API request, either as an `Authorization: Bearer <token>` header or via a Clerk-managed session cookie.
5. **Go backend middleware verifies JWT** -- an authentication middleware on the Go backend verifies the JWT signature using Clerk's public keys (via the Clerk Go SDK or the JWKS endpoint). Invalid or expired tokens are rejected with `401 Unauthorized`.
6. **Middleware extracts `clerk_user_id` and looks up local role** -- after verifying the JWT, the middleware extracts the `clerk_user_id` claim and queries the local `users` table to retrieve the user's role.
7. **Auto-provisioning (if no local record exists)** -- if no matching record is found in the `users` table, the middleware auto-creates a new record with the `'viewer'` role. Alternatively, if the system is configured to require admin pre-approval, the request is rejected with `403 Forbidden`.

## User Provisioning

### Creating Users

- **Clerk dashboard** -- admins create new users directly in the Clerk dashboard (useful for small deployments)
- **Clerk invitation API** -- admins send email invitations through Clerk's API, allowing users to set up their own credentials

### Role Assignment

- **First user** -- the first user to sign up is automatically assigned the `'admin'` role. This can also be configured via an environment variable (e.g., `HOMEPHOTOS_ADMIN_CLERK_ID=user_2abc...`) to pre-assign admin to a specific Clerk user ID.
- **Subsequent users** -- all users after the first receive the `'viewer'` role by default.
- **Promotion / demotion** -- admins can change any user's role through the HomePhotos admin UI, which calls the role update API endpoint.

## API Endpoints

### `GET /api/v1/auth/me`

Returns the current authenticated user's information, combining data from the Clerk JWT and the local `users` table.

**Authentication:** Required (any role)

**Response:**

```json
{
  "id": 1,
  "clerk_user_id": "user_2abc...",
  "role": "admin",
  "display_name": "Jane",
  "created_at": "2025-12-01T10:00:00Z",
  "last_login": "2026-03-31T18:30:00Z"
}
```

### `GET /api/v1/users`

Lists all users registered in the local `users` table.

**Authentication:** Required (admin only)

**Response:**

```json
{
  "users": [
    {
      "id": 1,
      "clerk_user_id": "user_2abc...",
      "role": "admin",
      "display_name": "Jane",
      "created_at": "2025-12-01T10:00:00Z",
      "last_login": "2026-03-31T18:30:00Z"
    },
    {
      "id": 2,
      "clerk_user_id": "user_3def...",
      "role": "viewer",
      "display_name": "Alex",
      "created_at": "2026-01-15T08:00:00Z",
      "last_login": "2026-03-30T12:00:00Z"
    }
  ]
}
```

### `PUT /api/v1/users/:id/role`

Updates a user's role.

**Authentication:** Required (admin only)

**Request body:**

```json
{
  "role": "admin"
}
```

**Response:**

```json
{
  "id": 2,
  "clerk_user_id": "user_3def...",
  "role": "admin",
  "display_name": "Alex",
  "created_at": "2026-01-15T08:00:00Z",
  "last_login": "2026-03-30T12:00:00Z"
}
```

**Errors:**

- `400 Bad Request` -- invalid role value (must be `'admin'` or `'viewer'`)
- `403 Forbidden` -- caller is not an admin
- `404 Not Found` -- user ID does not exist

## Security Considerations

| Concern | How It Is Handled |
|---|---|
| **Rate limiting** | Delegated to Clerk. Clerk applies rate limits to its authentication endpoints, protecting against brute-force attacks. |
| **CSRF protection** | Handled by Clerk's token mechanism. Clerk's session tokens are not vulnerable to CSRF when used as `Authorization` headers. If using cookies, Clerk's SDK includes CSRF protections. |
| **JWT expiry and refresh** | Handled by the Clerk SDK on the frontend. Short-lived JWTs are issued and automatically refreshed by the Clerk client before expiry. The backend only needs to verify the token; it does not manage refresh logic. |
| **Token storage** | Clerk's SDK manages token storage securely (httpOnly cookies or in-memory, depending on configuration). |
| **Secrets management** | The Clerk API secret key is stored as an environment variable on the backend, never committed to source control. |
