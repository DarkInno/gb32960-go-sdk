package gb32960

import (
	"crypto/rand"
	"testing"
)

func TestDeriveAESKey(t *testing.T) {
	t.Run("full token", func(t *testing.T) {
		token := make([]byte, 16)
		for i := range token {
			token[i] = byte(i)
		}
		key := DeriveAESKey(token)
		if len(key) != 16 {
			t.Fatalf("expected key length 16, got %d", len(key))
		}
		for i := range key {
			if key[i] != byte(i) {
				t.Errorf("key[%d] = %d, want %d", i, key[i], byte(i))
			}
		}
	})

	t.Run("short token", func(t *testing.T) {
		token := []byte("hello")
		key := DeriveAESKey(token)
		if len(key) != 16 {
			t.Fatalf("expected key length 16, got %d", len(key))
		}
		for i := 0; i < 5; i++ {
			if key[i] != token[i] {
				t.Errorf("key[%d] = %d, want %d", i, key[i], token[i])
			}
		}
		for i := 5; i < 16; i++ {
			if key[i] != 0 {
				t.Errorf("key[%d] expected zero-padding, got %d", i, key[i])
			}
		}
	})

	t.Run("empty token", func(t *testing.T) {
		key := DeriveAESKey(nil)
		if len(key) != 16 {
			t.Fatalf("expected key length 16, got %d", len(key))
		}
		for i := range key {
			if key[i] != 0 {
				t.Errorf("key[%d] expected 0, got %d", i, key[i])
			}
		}
	})
}

func TestAES128Roundtrip(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i + 1)
	}

	testCases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"one byte", []byte{0x42}},
		{"16 bytes (exact block)", make([]byte, 16)},
		{"17 bytes (partial block)", make([]byte, 17)},
		{"256 bytes", make([]byte, 256)},
		{"json data", []byte(`{"type":"realtime","vin":"TESTVIN1234567890","data":{"speed":60}}`)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.data) > 0 && tc.data[0] == 0 {
				copy(tc.data, make([]byte, len(tc.data)))
				for i := range tc.data {
					tc.data[i] = byte(i & 0xFF)
				}
			}

			encrypted, err := EncryptAES128(tc.data, key)
			if err != nil {
				t.Fatalf("encrypt error: %v", err)
			}

			if len(encrypted) == 0 && len(tc.data) > 0 {
				t.Error("encrypted data is empty but original was not")
			}

			decrypted, err := DecryptAES128(encrypted, key)
			if err != nil {
				t.Fatalf("decrypt error: %v", err)
			}

			if string(decrypted) != string(tc.data) {
				t.Errorf("roundtrip mismatch: got %v, want %v", decrypted, tc.data)
			}
		})
	}
}

func TestAES128LargeData(t *testing.T) {
	key := make([]byte, 16)
	_, _ = rand.Read(key)

	data := make([]byte, 65535)
	_, _ = rand.Read(data)

	encrypted, err := EncryptAES128(data, key)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := DecryptAES128(encrypted, key)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if len(decrypted) != len(data) {
		t.Errorf("length mismatch: got %d, want %d", len(decrypted), len(data))
	}
	for i := range data {
		if decrypted[i] != data[i] {
			t.Errorf("mismatch at byte %d", i)
			break
		}
	}
}

func TestDecryptAES128Errors(t *testing.T) {
	key := make([]byte, 16)

	t.Run("short key", func(t *testing.T) {
		_, err := DecryptAES128([]byte{0x00}, []byte("short"))
		if err == nil {
			t.Error("expected error for short key")
		}
	})

	t.Run("invalid ciphertext", func(t *testing.T) {
		_, err := DecryptAES128([]byte{0x00, 0x01, 0x02}, key)
		if err == nil {
			t.Error("expected error for non-block-aligned ciphertext")
		}
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		data := []byte("hello world test")
		encrypted, _ := EncryptAES128(data, key)
		encrypted[0] ^= 0xFF
		decrypted, err := DecryptAES128(encrypted, key)
		if err != nil {
			return
		}
		if string(decrypted) == string(data) {
			t.Error("corrupted ciphertext roundtripped correctly, should not match")
		}
	})
}
