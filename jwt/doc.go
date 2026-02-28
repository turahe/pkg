/*
Package jwt provides JWT token generation and validation with configurable signing:
HS256 (symmetric), RS256 (RSA, default), or ES256 (ECDSA). Secret or keys are loaded from
config (env, file paths, or embedded PEM bytes via config.Server.JWTPrivateKeyPEM/JWTPublicKeyPEM).

Role in architecture:
  - Infrastructure: used by auth middleware and handlers; reads config.Server (Secret, key paths or embedded PEM) and expiry.

Responsibilities:
  - NewManager: all-in-one sign + verify; loads keys from config (env, file, or embed); returns error instead of panic.
  - NewSigner: signing only (private key or secret); use for auth/login services that issue tokens.
  - NewVerifier: verification only (public key or secret); use for API/gateway services that only validate tokens.
  - TokenVerifier interface: implemented by *Manager and *Verifier; pass to AuthMiddleware so either can be used.
  - Manager/Signer: GenerateToken, GenerateTokenWithExpiry, GenerateRefreshToken, GenerateImpersonationToken.
  - Manager/Verifier: ValidateToken.
  - ComparePassword: bcrypt. GetCurrentUserUUID: read user_id from Gin context.

Constraints:
  - Default algorithm is RS256; set JWT_SIGNING_ALGORITHM=HS256 for symmetric secret.
  - For RS256/ES256, provide keys via JWT_PRIVATE_KEY and JWT_PUBLIC_KEY (path or inline PEM), or set config.Server.JWTPrivateKeyPEM and JWTPublicKeyPEM (e.g. from //go:embed) before calling NewManager/NewSigner/NewVerifier.
  - No global state; create Manager, Signer, or Verifier via NewManager/NewSigner/NewVerifier(ctx, config).

This package must NOT:
  - Contain use-case logic; only token and password operations.
*/
package jwt
