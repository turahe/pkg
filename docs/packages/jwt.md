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
token, err := m.GenerateToken(userID)
refresh, err := m.GenerateRefreshToken(userID)
imp, err := m.GenerateImpersonationToken(adminID, "admin", targetID, 15*time.Minute) // TTL capped at 30m
claims, err := m.ValidateToken(tokenString)
```

Token types: `TokenTypeAccess`, `TokenTypeRefresh`, `TokenTypeImpersonation`.

Also: `ComparePassword` (bcrypt), `GetCurrentUserUUID(c *gin.Context)`.

## Key loading

1. `JWT_SIGNING_ALGORITHM` — default `RS256` (`ES256` also supported; `HS256` rejected)
2. `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` (file path **or** inline PEM containing `-----BEGIN`)
3. Or set `Server.JWTPrivateKeyPEM` / `JWTPublicKeyPEM` (e.g. `//go:embed`) before `New*`

Optional claims: `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_KEY_ID`.

## See also

- [middlewares — Auth](middlewares.md)
- [config](config.md)
- [Environment — Server & JWT](../environment.md#server--jwt)
