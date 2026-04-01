# Feature: User Management

## Overview

HomePhotos uses built-in username/password authentication with bcrypt password hashing and JWT tokens. There is no external authentication service. The backend handles user registration, login, password verification, and JWT issuance. The frontend provides its own login and signup forms.

This keeps the application fully self-contained -- it can run on a home network with no internet connectivity required for authentication.

## Authentication Design

### Password Storage

Passwords are hashed using bcrypt with a cost factor of 12 before being stored in the database. Plaintext passwords are never stored or logged.

### JWT Tokens

On successful login, the backend issues a JWT signed with HMAC-SHA256 using a server-side secret (`HOMEPHOTOS_JWT_SECRET` environment variable). The JWT contains:

- `sub` -- the user's local ID
- `role` -- the user's role (`admin` or `viewer`)
- `exp` -- expiration timestamp (default: 24 hours from issuance, configurable)

The frontend stores the JWT and sends it as an `Authorization: Bearer <token>` header on every API request.

### Frontend Integration

The SvelteKit frontend provides its own login and registration forms:

- **Login form** -- username and password fields. Submits credentials to `POST /api/v1/auth/login`.
- **Registration form** -- username, password, and optional email fields. Submits to `POST /api/v1/auth/register`. Only shown when registration is open.
- **User menu** -- displays the signed-in user's name with a sign-out option that clears the stored token.

### Backend Integration

The Go backend handles all authentication concerns:

1. **Registration** -- validates input, hashes the password with bcrypt, stores the user record.
2. **Login** -- looks up the user by username, verifies the password against the stored bcrypt hash, issues a JWT on success.
3. **Request authentication** -- middleware extracts the JWT from the `Authorization` header, verifies the signature with the server secret, extracts the user ID and role from claims.

## User Model

### SQLite Schema

```sql
CREATE TABLE users (
    id              INTEGER PRIMARY KEY,
    username        TEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    email           TEXT,
    role            TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'viewer')),
    display_name    TEXT,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login      DATETIME
);
```

| Column | Type | Description |
|---|---|---|
| `id` | `INTEGER PRIMARY KEY` | Local auto-incrementing ID |
| `username` | `TEXT UNIQUE NOT NULL` | Unique login name |
| `password_hash` | `TEXT NOT NULL` | bcrypt hash of the user's password |
| `email` | `TEXT` | Optional email address |
| `role` | `TEXT NOT NULL` | Either `'admin'` or `'viewer'`. Defaults to `'viewer'`. |
| `display_name` | `TEXT` | Optional display name |
| `created_at` | `DATETIME` | When the record was created |
| `last_login` | `DATETIME` | Updated on each successful login |

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

1. **User visits HomePhotos** -- the SvelteKit app loads and shows the login form if no valid JWT is present.
2. **User submits credentials** -- the frontend sends the username and password to `POST /api/v1/auth/login`.
3. **Backend verifies password** -- the backend looks up the user by username and verifies the submitted password against the stored bcrypt hash.
4. **Backend issues JWT** -- on successful verification, the backend signs a JWT containing the user ID and role, and returns it in the response body.
5. **Frontend stores token** -- the frontend stores the JWT (in memory or localStorage) and attaches it to every subsequent API request as an `Authorization: Bearer <token>` header.
6. **Backend middleware verifies JWT** -- on each API request, middleware extracts the JWT from the `Authorization` header, verifies the HMAC-SHA256 signature using the server secret, checks expiration, and extracts the user ID and role from claims. Invalid or expired tokens are rejected with `401 Unauthorized`.
7. **Request proceeds with user context** -- the authenticated user's ID and role are attached to the request context for downstream handlers.

## User Registration

### First User

The first user to register is automatically assigned the `admin` role. All subsequent users receive the `viewer` role by default.

### Registration Control

Registration can be opened or closed via the `HOMEPHOTOS_REGISTRATION_OPEN` environment variable (default: `true`). When set to `false`, the registration endpoint returns `403 Forbidden` and the frontend hides the registration form. New users can only be added by an admin at that point.

### Role Assignment

- **First user** -- automatically gets the `admin` role.
- **Subsequent users** -- all receive the `viewer` role by default.
- **Promotion / demotion** -- admins can change any user's role through the HomePhotos admin UI, which calls the role update API endpoint.

## API Endpoints

### `POST /api/v1/auth/register`

Creates a new user account.

**Authentication:** None required

**Request body:**

```json
{
  "username": "jane",
  "password": "correct-horse-battery-staple",
  "email": "jane@example.com"
}
```

The `email` field is optional.

**Response** `201 Created`:

```json
{
  "id": 1,
  "username": "jane",
  "role": "admin",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

The first registered user receives the `admin` role. A JWT is returned immediately so the user is signed in after registration.

**Errors:**

- `400 Bad Request` -- missing username or password, or username already taken
- `403 Forbidden` -- registration is closed (`HOMEPHOTOS_REGISTRATION_OPEN=false`)

### `POST /api/v1/auth/login`

Authenticates a user and returns a JWT.

**Authentication:** None required

**Request body:**

```json
{
  "username": "jane",
  "password": "correct-horse-battery-staple"
}
```

**Response** `200 OK`:

```json
{
  "id": 1,
  "username": "jane",
  "role": "admin",
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Errors:**

- `401 Unauthorized` -- invalid username or password (same error for both to avoid user enumeration)

### `GET /api/v1/auth/me`

Returns the current authenticated user's information.

**Authentication:** Required (any role)

**Response:**

```json
{
  "id": 1,
  "username": "jane",
  "role": "admin",
  "display_name": "Jane",
  "email": "jane@example.com",
  "created_at": "2025-12-01T10:00:00Z",
  "last_login": "2026-03-31T18:30:00Z"
}
```

### `GET /api/v1/users`

Lists all users.

**Authentication:** Required (admin only)

**Response:**

```json
{
  "users": [
    {
      "id": 1,
      "username": "jane",
      "role": "admin",
      "display_name": "Jane",
      "created_at": "2025-12-01T10:00:00Z",
      "last_login": "2026-03-31T18:30:00Z"
    },
    {
      "id": 2,
      "username": "alex",
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
  "username": "alex",
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
| **Password storage** | Passwords are hashed with bcrypt using a cost factor of 12. Plaintext passwords are never stored or logged. |
| **JWT signing** | JWTs are signed with HMAC-SHA256 using the `HOMEPHOTOS_JWT_SECRET` environment variable. The secret must be a strong random string. |
| **JWT expiry** | Tokens expire after 24 hours by default. Expired tokens are rejected by the middleware. Users must re-authenticate to get a new token. |
| **Token storage** | The frontend stores the JWT in memory or localStorage. The token is cleared on sign-out. |
| **User enumeration** | The login endpoint returns the same error message for invalid username and invalid password to prevent user enumeration. |
| **Secrets management** | `HOMEPHOTOS_JWT_SECRET` is stored as an environment variable, never committed to source control. |
| **Brute force** | For a home network app behind Tailscale, brute force is a low risk. The bcrypt cost factor provides inherent rate limiting on password verification. |
