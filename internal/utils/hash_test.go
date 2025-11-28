package utils

import (
	"bytes"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name    string
		src     []byte
		key     string
		wantErr bool
	}{
		{
			name:    "valid hash",
			src:     []byte("test data"),
			key:     "secret-key",
			wantErr: false,
		},
		{
			name:    "empty key",
			src:     []byte("test data"),
			key:     "",
			wantErr: true,
		},
		{
			name:    "empty data",
			src:     []byte(""),
			key:     "secret-key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Hash(tt.src, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 {
				t.Errorf("Hash() returned empty result")
			}
		})
	}
}

func TestHashDeterministic(t *testing.T) {
	data := []byte("test data")
	key := "secret-key"

	hash1, err1 := Hash(data, key)
	hash2, err2 := Hash(data, key)

	if err1 != nil || err2 != nil {
		t.Fatalf("Hash() failed: err1=%v, err2=%v", err1, err2)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Errorf("Hash() is not deterministic: %x != %x", hash1, hash2)
	}
}

func TestHashDifferentKeys(t *testing.T) {
	data := []byte("test data")
	key1 := "secret-key-1"
	key2 := "secret-key-2"

	hash1, err1 := Hash(data, key1)
	hash2, err2 := Hash(data, key2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Hash() failed: err1=%v, err2=%v", err1, err2)
	}

	if bytes.Equal(hash1, hash2) {
		t.Errorf("Hash() produced same hash for different keys")
	}
}

func TestHashLength(t *testing.T) {
	data := []byte("test data for length check")
	key := "shared-secret-key"

	hash1, err := Hash(data, key)
	if err != nil {
		t.Fatalf("Hash() failed: %v", err)
	}

	if len(hash1) != 32 {
		t.Errorf("Hash() length = %d, want 32 (SHA256 produces 32 bytes)", len(hash1))
	}
}
