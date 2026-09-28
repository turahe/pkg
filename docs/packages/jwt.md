# jwt

JWT generation and validation with RS256 (default) or ES256. Keys from env (path or inline PEM), or embedded PEM bytes on `config.Server`. HS256 is not supported.

**Import:** `github.com/turahe/pkg/jwt`

## Role

Infrastructure for auth middleware and login handlers. No use-case logic; no global singleton — construct via `New*`.

## Constructors

| Constructor | Use when |
|-------------|----------|
| `NewManager(ctx, cfg)` | Sign + verify (monolith / auth service) |
| `NewSigner(ctx, cfg)` | Issue tokens only |
| `NewVerifier(ctx, cfg)` | Validate only (API / gateway) |

`TokenVerifier` is implemented by `*Manager` and `*Verifier` — pass to `middlewares.AuthMiddleware(verifier)`.

## Tokens

```go
m, err := jwt.NewManager(ctx, cfg)
token, err := m.GenerateToken(userID, "users")   // actor_type=user (users/User table)
adminTok, err := m.GenerateToken(adminID, "admins") // actor_type=admin
svcTok, err := m.GenerateToken(svcID)              // actor_type=service (empty)
refresh, err := m.GenerateRefreshToken(userID, "users")
imp, err := m.GenerateImpersonationToken(adminID, "admin", targetID, 15*time.Minute) // actor_type=user
claims, err := m.ValidateToken(tokenString)             // signature/expiry only, any token_type
claims, err = jwt.ValidateAccessToken(m, tokenString)   // access or impersonation only
claims, err = jwt.ValidateRefreshToken(m, refreshToken) // refresh only (use in /refresh endpoint)
```

Token types: `TokenTypeAccess`, `TokenTypeRefresh`, `TokenTypeImpersonation`. `ValidateAccessToken` / `ValidateRefreshToken` return `ErrWrongTokenType` when the type does not match; `AuthMiddleware` uses `ValidateAccessToken`, so refresh tokens are rejected as bearer credentials.

Actor types (`actor_type` claim via `ResolveActorType`):

| Input (type or table) | `actor_type` |
|-----------------------|--------------|
| `users` / `User` / `user` | `user` |
| `admins` / `admin` | `admin` |
| empty / omitted | `service` |
| `system` | `system` |
| `service` | `service` |

Also: `ComparePassword` (bcrypt), `GetCurrentUserUUID(c *gin.Context)`, `GetActorType(c *gin.Context)`.

## Key loading

1. `JWT_SIGNING_ALGORITHM` — default `RS256` (`ES256` also supported; `HS256` rejected)
2. `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` (file path **or** inline PEM containing `-----BEGIN`)
3. Or set `Server.JWTPrivateKeyPEM` / `JWTPublicKeyPEM` (e.g. `//go:embed`) before `New*`

Optional claims: `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_KEY_ID`.

## See also

- [middlewares — Auth](middlewares.md)
- [config](config.md)
- [Environment — Server & JWT](../environment.md#server--jwt)
