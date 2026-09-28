/*
Package crypto provides password hashing behind a driver pattern: bcrypt (golang.org/x/crypto/bcrypt),
argon (Argon2i) and argon2id (golang.org/x/crypto/argon2).

Role in architecture:
  - Infrastructure utility: no business logic; used by auth or user flows that need to hash/compare passwords.

Responsibilities:
  - Hasher: Driver / Hash / Check / NeedsRehash. Built with NewHasher(driver, HasherOptions) or
    NewHasherFromConfig(config.HashingConfiguration). Drivers: DriverBcrypt ("bcrypt", default),
    DriverArgon ("argon" = Argon2i), DriverArgon2id ("argon2id"); unknown names return ErrUnknownDriver.
  - Setup: build the default hasher from HASH_DRIVER / HASH_BCRYPT_ROUNDS / HASH_ARGON_* (config.GetConfig().Hashing).
    SetDefault / Default replace or read it; without Setup the default is bcrypt at DefaultCost.
  - Hash / NeedsRehash: package-level helpers using Default().
  - ComparePassword and Hasher.Check: constant-time check that accepts every supported format
    ($argon2id$, $argon2i$, bcrypt) regardless of the configured driver, so switching drivers keeps old hashes valid.
  - HashArgon2id / HashArgon2idWithParams / VerifyArgon2id / IsArgon2idHash: direct argon2id helpers.
    Argon2 hashes are PHC strings ($argon2id$v=19$m=..,t=..,p=..$salt$hash); DefaultArgon2Params follows
    the OWASP baseline (19 MiB, 2 iterations, parallelism 1).
  - HashAndSalt: hash with bcrypt.DefaultCost; log and return empty string on error.

Constraints:
  - Decoded argon2 parameters are bounded (memory <= 1 GiB, iterations <= 64) so stored values cannot force huge work.
  - Depends on config (Setup) and logger (error logging).

This package must NOT:
  - Store or retrieve passwords; only hash and compare.
*/
package crypto
