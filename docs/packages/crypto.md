# crypto

Password hashing with selectable drivers: `bcrypt` (default), `argon` (Argon2i) and `argon2id`.

**Import:** `github.com/turahe/pkg/crypto`

## Hash drivers

| Driver | Constant | Algorithm |
|--------|----------|-----------|
| `bcrypt` | `crypto.DriverBcrypt` | bcrypt (default) |
| `argon` | `crypto.DriverArgon` | Argon2i |
| `argon2id` | `crypto.DriverArgon2id` | Argon2id (recommended for new apps) |

Configure via env (see [environment](../environment.md#password-hashing)) and call `Setup` once at startup:

```go
crypto.Setup() // reads HASH_DRIVER, HASH_BCRYPT_ROUNDS, HASH_ARGON_MEMORY/TIME/THREADS

hash, err := crypto.Hash([]byte(plain))              // uses the configured driver
ok := crypto.ComparePassword(hash, []byte(plain))    // accepts any supported format
stale := crypto.NeedsRehash(hash)                    // true if not produced by the current driver/params
```

Or build a hasher explicitly:

```go
h, err := crypto.NewHasher(crypto.DriverArgon2id, crypto.HasherOptions{
    Argon: crypto.Argon2Params{Memory: 64 * 1024, Iterations: 3, Parallelism: 4}, // zero fields = defaults
})
// crypto.NewHasher(crypto.DriverBcrypt, crypto.HasherOptions{BcryptCost: 12})

hash, err := h.Hash([]byte(plain))
ok := h.Check(hash, []byte(plain))
if h.NeedsRehash(hash) { /* re-hash and persist */ }

crypto.SetDefault(h) // make it the package default (nil restores bcrypt)
```

Unknown driver names return `crypto.ErrUnknownDriver`; out-of-range bcrypt cost or Argon2 params are rejected when the hasher is built.

## Switching drivers

`Check` / `ComparePassword` verify bcrypt, Argon2i and Argon2id hashes regardless of the configured driver, so changing `HASH_DRIVER` never locks users out. Upgrade hashes on login:

```go
if crypto.ComparePassword(user.PasswordHash, []byte(plain)) && crypto.NeedsRehash(user.PasswordHash) {
    if h, err := crypto.Hash([]byte(plain)); err == nil {
        user.PasswordHash = h // persist
    }
}
```

`NeedsRehash` is true for hashes from another driver, malformed values, a different bcrypt cost, or different Argon2 memory/iterations/parallelism/key length.

## Low-level helpers

```go
// argon2id (PHC string: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>)
hash, err := crypto.HashArgon2id([]byte(plain))
hash, err = crypto.HashArgon2idWithParams([]byte(plain), crypto.Argon2Params{
    Memory: 64 * 1024, Iterations: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32, // RFC 9106 profile
})
ok, err := crypto.VerifyArgon2id(hash, []byte(plain)) // err = ErrInvalidArgon2Hash if malformed

// bcrypt at DefaultCost, ignores HASH_DRIVER
hash := crypto.HashAndSalt([]byte(plain)) // empty string on error
```

`DefaultArgon2Params` follows the OWASP baseline (19 MiB memory, 2 iterations, parallelism 1, 16-byte salt, 32-byte key). Argon2 hashes are interoperable with other PHC-compatible implementations (e.g. argon2-cffi, libsodium-style encoders). When decoding, parameters are bounded (memory ≤ 1 GiB, iterations ≤ 64, encoded length ≤ 512) so a corrupted stored value cannot force excessive work.

## Constraints

- Does not store or retrieve passwords.
- No business rules — hash/compare only.
