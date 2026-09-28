package crypto

import (
	"errors"
	"strings"
	"testing"
)

// Generated with argon2-cffi (reference C implementation) for "correct horse battery staple".
const (
	refArgon2OWASP  = "$argon2id$v=19$m=19456,t=2,p=1$kqmDUm8cSsDyBbPFmdGHwg$dvWRrJ6Jir076RAMuEwf+dvaaRIQsJZUmjcKHJkDTNQ"
	refArgon2RFC    = "$argon2id$v=19$m=65536,t=3,p=4$9d8i2WfjlAw0ynSfWgHO0w$Y4z/ssvAxvz6PcaP55ZRdIUfmwww0DDQpWFkH2rb2OQ"
	refArgon2Secret = "correct horse battery staple"
)

func TestHashArgon2id_Format(t *testing.T) {
	hash, err := HashArgon2id([]byte("password123"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("hash = %q, want PHC prefix with default params", hash)
	}
	hash2, err := HashArgon2id([]byte("password123"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	if hash == hash2 {
		t.Error("same input should produce different hashes (random salt)")
	}
}

func TestVerifyArgon2id_RoundTrip(t *testing.T) {
	hash, err := HashArgon2id([]byte("secret"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	if ok, err := VerifyArgon2id(hash, []byte("secret")); err != nil || !ok {
		t.Errorf("VerifyArgon2id(correct) = %v, %v; want true, nil", ok, err)
	}
	if ok, err := VerifyArgon2id(hash, []byte("wrong")); err != nil || ok {
		t.Errorf("VerifyArgon2id(wrong) = %v, %v; want false, nil", ok, err)
	}
}

func TestVerifyArgon2id_ReferenceVectors(t *testing.T) {
	for _, h := range []string{refArgon2OWASP, refArgon2RFC} {
		ok, err := VerifyArgon2id(h, []byte(refArgon2Secret))
		if err != nil || !ok {
			t.Errorf("VerifyArgon2id(%q) = %v, %v; want true, nil", h, ok, err)
		}
		if ok, _ := VerifyArgon2id(h, []byte("Correct horse battery staple")); ok {
			t.Errorf("VerifyArgon2id(%q) accepted wrong password", h)
		}
	}
}

func TestVerifyArgon2id_Malformed(t *testing.T) {
	valid := refArgon2OWASP
	cases := map[string]string{
		"empty":             "",
		"bcrypt":            "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ01234",
		"argon2i variant":   strings.Replace(valid, "argon2id", "argon2i", 1),
		"wrong version":     strings.Replace(valid, "v=19", "v=16", 1),
		"trailing version":  strings.Replace(valid, "v=19", "v=19x", 1),
		"missing field":     strings.Replace(valid, "$kqmDUm8cSsDyBbPFmdGHwg", "", 1),
		"bad params":        strings.Replace(valid, "m=19456,t=2,p=1", "m=abc,t=2,p=1", 1),
		"trailing params":   strings.Replace(valid, "p=1", "p=1,x=2", 1),
		"zero parallelism":  strings.Replace(valid, "p=1", "p=0", 1),
		"huge memory":       strings.Replace(valid, "m=19456", "m=4194304", 1),
		"huge iterations":   strings.Replace(valid, "t=2", "t=1000", 1),
		"bad salt base64":   strings.Replace(valid, "kqmDUm8cSsDyBbPFmdGHwg", "!!!!", 1),
		"short salt":        strings.Replace(valid, "kqmDUm8cSsDyBbPFmdGHwg", "YWJj", 1),
		"bad key base64":    valid[:len(valid)-4] + "!!!!",
		"oversized encoded": valid + strings.Repeat("A", maxArgon2EncodedLength),
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			ok, err := VerifyArgon2id(h, []byte(refArgon2Secret))
			if ok || !errors.Is(err, ErrInvalidArgon2Hash) {
				t.Errorf("VerifyArgon2id = %v, %v; want false, ErrInvalidArgon2Hash", ok, err)
			}
		})
	}
}

func TestHashArgon2idWithParams_Validation(t *testing.T) {
	bad := map[string]Argon2Params{
		"zero memory":        {Memory: 0, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32},
		"memory below lanes": {Memory: 8, Iterations: 1, Parallelism: 4, SaltLength: 16, KeyLength: 32},
		"zero iterations":    {Memory: 64, Iterations: 0, Parallelism: 1, SaltLength: 16, KeyLength: 32},
		"zero parallelism":   {Memory: 64, Iterations: 1, Parallelism: 0, SaltLength: 16, KeyLength: 32},
		"short salt":         {Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 4, KeyLength: 32},
		"short key":          {Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 8},
	}
	for name, p := range bad {
		if _, err := HashArgon2idWithParams([]byte("x"), p); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}

	custom := Argon2Params{Memory: 64, Iterations: 1, Parallelism: 2, SaltLength: 8, KeyLength: 16}
	hash, err := HashArgon2idWithParams([]byte("x"), custom)
	if err != nil {
		t.Fatalf("HashArgon2idWithParams: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=64,t=1,p=2$") {
		t.Errorf("hash = %q, want custom params encoded", hash)
	}
	if ok, err := VerifyArgon2id(hash, []byte("x")); err != nil || !ok {
		t.Errorf("VerifyArgon2id(custom) = %v, %v", ok, err)
	}
}

func TestComparePassword_Argon2idAndBcrypt(t *testing.T) {
	argonHash, err := HashArgon2id([]byte("secret"))
	if err != nil {
		t.Fatalf("HashArgon2id: %v", err)
	}
	bcryptHash := HashAndSalt([]byte("secret"))

	for name, h := range map[string]string{"argon2id": argonHash, "bcrypt": bcryptHash, "argon2id reference": refArgon2OWASP} {
		plain := "secret"
		if name == "argon2id reference" {
			plain = refArgon2Secret
		}
		if !ComparePassword(h, []byte(plain)) {
			t.Errorf("%s: ComparePassword(correct) = false", name)
		}
		if ComparePassword(h, []byte("wrong")) {
			t.Errorf("%s: ComparePassword(wrong) = true", name)
		}
	}
	if ComparePassword("$argon2id$garbage", []byte("secret")) {
		t.Error("malformed argon2id hash must not verify")
	}
}
