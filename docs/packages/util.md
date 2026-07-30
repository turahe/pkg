# util

Generic helpers with no database, HTTP, or config dependencies (stdlib + phonenumbers).

**Import:** `github.com/turahe/pkg/util`

## API

```go
util.IsEmpty(value)                              // nil, "", empty slice/map, zero values
util.InAnySlice(haystack, needle)                // generic comparable
util.RemoveDuplicates(haystack)                  // preserve first-seen order
util.FormatPhoneNumber(phone, countryCode)       // E.164; nil on failure
util.FormatCurrency(amount, currencyCode)        // locale-aware currency string
```

## Constraints

- Must not depend on infrastructure packages.
- No business rules — formatting and collection helpers only.

## See also

- [Architecture — Shared](../architecture.md#layers)
