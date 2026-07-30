# crypto

bcrypt password hashing helpers.

**Import:** `github.com/turahe/pkg/crypto`

## API

```go
hash := crypto.HashAndSalt([]byte(plain))           // empty string on error
ok := crypto.ComparePassword(hash, []byte(plain))   // false if hash too short / empty
```

Uses `bcrypt.MinCost` only (fixed). `ComparePassword` returns `false` without calling bcrypt when the stored hash is empty or shorter than 60 characters (avoids noisy bcrypt errors).

## Constraints

- Does not store or retrieve passwords.
- No business rules — hash/compare only.

## See also

- [jwt.ComparePassword](jwt.md) (related helper on jwt package)
