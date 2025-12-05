/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package encrypt

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/asgardeo/thunder/internal/system/config"

	"github.com/stretchr/testify/require"
)

// Mock config for testing.
type MockThunderRuntime struct {
	Config struct {
		Crypto struct {
			Key string
		}
	}
}

const (
	expectedPath = "/home/thunder/config/crypto.key"
)

func TestInitEncryptionServiceCryptoFilePathNotFound(t *testing.T) {
	mockFileReader := func(name string) ([]byte, error) { return nil, nil }
	mockPathJoiner := func(elem ...string) string { return "/mock/path" }

	mockConfig := &config.Config{
		Security: config.SecurityConfig{
			CryptoFile: "",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/home/thunder", mockConfig)
	require.NoError(t, err)

	service, err := initEncryptionService(
		mockFileReader,
		mockPathJoiner,
	)

	config.ResetThunderRuntime()

	require.Error(t, err)
	require.Contains(t, err.Error(), "crypto key file path not found in configs")
	require.Nil(t, service)
}

func TestInitEncryptionServiceCryptoFileReadError(t *testing.T) {
	expectedError := errors.New("permission denied")

	mockFileReader := func(name string) ([]byte, error) {
		require.Equal(t, expectedPath, name)
		return nil, expectedError
	}

	mockPathJoiner := func(elem ...string) string {
		require.Equal(t, "/home/thunder", elem[0])
		require.Equal(t, "config/crypto.key", elem[1])
		return expectedPath
	}

	mockConfig := &config.Config{
		Security: config.SecurityConfig{
			CryptoFile: "config/crypto.key",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/home/thunder", mockConfig)
	require.NoError(t, err)

	service, err := initEncryptionService(
		mockFileReader,
		mockPathJoiner,
	)

	config.ResetThunderRuntime()

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read crypto key file at path")
	require.Contains(t, err.Error(), expectedPath)
	require.ErrorIs(t, err, expectedError)
	require.Nil(t, service)
}

func TestInitEncryptionServiceInvalidHexInCryptoFile(t *testing.T) {
	invalidHexData := []byte("INVALID_HEX_STRING")

	mockFileReader := func(name string) ([]byte, error) {
		return invalidHexData, nil
	}

	mockPathJoiner := func(elem ...string) string {
		return expectedPath
	}
	mockConfig := &config.Config{
		Security: config.SecurityConfig{
			CryptoFile: "config/crypto.key",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/home/thunder", mockConfig)
	require.NoError(t, err)

	service, err := initEncryptionService(
		mockFileReader,
		mockPathJoiner,
	)

	config.ResetThunderRuntime()

	require.Error(t, err)
	require.Contains(t, err.Error(), "error while reading crypto key file at path")
	require.Contains(t, err.Error(), expectedPath)
	require.Contains(t, err.Error(), "invalid byte")
	require.Nil(t, service)
}

func TestInitEncryptionServiceSuccess(t *testing.T) {
	validKey, _ := generateRandomKey(defaultKeySize)
	validHexData := []byte(hex.EncodeToString(validKey))

	mockFileReader := func(name string) ([]byte, error) {
		return validHexData, nil
	}

	mockPathJoiner := func(elem ...string) string {
		return expectedPath
	}

	mockConfig := &config.Config{
		Security: config.SecurityConfig{
			CryptoFile: "config/crypto.key",
		},
	}
	config.ResetThunderRuntime()
	err := config.InitializeThunderRuntime("/home/thunder", mockConfig)
	require.NoError(t, err)

	service, err := initEncryptionService(
		mockFileReader,
		mockPathJoiner,
	)

	config.ResetThunderRuntime()

	require.NoError(t, err)
	require.NotNil(t, service)
	require.Equal(t, validKey, service.Key)
	require.Equal(t, getKeyID(validKey), service.Kid)
}
func TestEncryptionService(t *testing.T) {
	// Generate a random key
	key, err := generateRandomKey(defaultKeySize)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	t.Logf("Generated random key: %x", key)

	// Create crypto service
	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("Failed to create crypto service: %v", err)
	}

	// Test data
	original := "This is a secret message that needs encryption!"

	// Encrypt
	encrypted, err := service.EncryptString(original)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	t.Logf("Encrypted data: %x", encrypted)

	// Decrypt
	decrypted, err := service.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	t.Logf("Decrypted data: %s", decrypted)

	// Verify
	if decrypted != original {
		t.Errorf("Decryption result doesn't match original. Got %q, want %q", decrypted, original)
	}
}

func TestTampering(t *testing.T) {
	// Generate a random key
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	// Encrypt some data
	original := "Protected data"
	encrypted, _ := service.EncryptString(original)

	// Parse the JSON to get the encrypted data structure
	var encData EncryptedData
	err := json.Unmarshal([]byte(encrypted), &encData)
	if err != nil {
		t.Fatalf("Failed to parse encrypted JSON: %v", err)
	}

	// Tamper with the ciphertext field
	cipherBytes := []byte(encData.Ciphertext)
	if len(cipherBytes) > 10 {
		cipherBytes[len(cipherBytes)-5] ^= 0x01 // Flip a bit in the base64 encoded ciphertext
	}
	encData.Ciphertext = string(cipherBytes)

	// Re-encode to JSON
	tamperedJSON, err := json.Marshal(encData)
	if err != nil {
		t.Fatalf("Failed to marshal tampered data: %v", err)
	}

	// Attempt to decrypt tampered data
	out, err := service.DecryptString(string(tamperedJSON))
	if err == nil {
		t.Error("Expected decryption of tampered data to fail, but it succeeded", out)
	}
}

func TestEncryptedObjectFormat(t *testing.T) {
	// Generate a random key
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	// Encrypt some data
	original := "Data to encrypt"
	encrypted, _ := service.EncryptString(original)

	// Parse the JSON to verify structure
	var encData EncryptedData
	err := json.Unmarshal([]byte(encrypted), &encData)
	if err != nil {
		t.Fatalf("Failed to parse encrypted JSON: %v", err)
	}

	// Verify the structure
	if encData.Algorithm != AESGCM {
		t.Errorf("Expected algorithm %s, got %s", AESGCM, encData.Algorithm)
	}
	if encData.Ciphertext == "" {
		t.Error("Ciphertext should not be empty")
	}
	if encData.KeyID != getKeyID(key) {
		t.Error("KeyID should match the expected value")
	}
}

func TestEncryptDecryptCycle(t *testing.T) {
	// Generate a key
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	// Test various data types
	testCases := []string{
		"",                               // Empty string
		"Hello World",                    // Simple text
		"特殊文字列",                          // Unicode characters
		"123456789012345678901234567890", // Long string
		`{"name":"John","age":30}`,       // JSON string
	}

	for _, tc := range testCases {
		encrypted, err := service.EncryptString(tc)
		if err != nil {
			t.Errorf("Failed to encrypt %q: %v", tc, err)
			continue
		}

		decrypted, err := service.DecryptString(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt %q: %v", tc, err)
			continue
		}

		if decrypted != tc {
			t.Errorf("Decryption result doesn't match original. Got %q, want %q", decrypted, tc)
		}
		t.Logf("Decryption successful. Decrypted data: %q", decrypted)
	}
}

func TestDifferentKeysEncryption(t *testing.T) {
	// Generate two different keys
	key1, _ := generateRandomKey(defaultKeySize)
	key2, _ := generateRandomKey(defaultKeySize)

	service1, _ := NewEncryptionService(key1)
	service2, _ := NewEncryptionService(key2)

	// Encrypt with first service
	original := "Secret message"
	encrypted, err := service1.EncryptString(original)
	if err != nil {
		t.Fatalf("Encryption with first key failed: %v", err)
	}

	// Try to decrypt with second service (should fail)
	_, err = service2.DecryptString(encrypted)
	if err == nil {
		t.Error("Expected decryption with different key to fail, but it succeeded")
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	service := &EncryptionService{
		Key: []byte("short"),
		Kid: "kid",
	}

	_, err := service.Encrypt([]byte("data"))
	require.Error(t, err)
}

func TestDecryptInvalidJSON(t *testing.T) {
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	_, err := service.Decrypt("not-json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid data format")
}

func TestDecryptUnsupportedAlgorithm(t *testing.T) {
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	payload := EncryptedData{
		Algorithm:  "RSA",
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("cipher")),
		KeyID:      service.Kid,
	}
	raw, _ := json.Marshal(payload)

	_, err := service.Decrypt(string(raw))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported algorithm")
}

func TestDecryptInvalidBase64(t *testing.T) {
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	payload := EncryptedData{
		Algorithm:  AESGCM,
		Ciphertext: "###invalid###",
		KeyID:      service.Kid,
	}
	raw, _ := json.Marshal(payload)

	_, err := service.Decrypt(string(raw))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid payload encoding")
}

func TestDecryptCiphertextTooShort(t *testing.T) {
	key, _ := generateRandomKey(defaultKeySize)
	service, _ := NewEncryptionService(key)

	payload := EncryptedData{
		Algorithm:  AESGCM,
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
		KeyID:      service.Kid,
	}
	raw, _ := json.Marshal(payload)

	_, err := service.Decrypt(string(raw))
	require.Error(t, err)
	require.EqualError(t, err, "ciphertext too short")
}

func TestDecryptWithInvalidKeyLength(t *testing.T) {
	service := &EncryptionService{
		Key: []byte("short"),
		Kid: "kid",
	}

	payload := EncryptedData{
		Algorithm:  AESGCM,
		Ciphertext: base64.StdEncoding.EncodeToString([]byte("ciphertext-with-nonce")),
		KeyID:      "kid",
	}
	raw, _ := json.Marshal(payload)

	_, err := service.Decrypt(string(raw))
	require.Error(t, err)
}

func TestNondefaultKeySize(t *testing.T) {
	// Test various key sizes
	testCases := []int{16, 24} // 128, 192 bits
	// Test data
	original := "This is a secret message that needs encryption!"
	for _, size := range testCases {
		key, _ := generateRandomKey(size)
		service, _ := NewEncryptionService(key)

		encrypted, err := service.EncryptString(original)
		if err != nil {
			t.Errorf("Failed to encrypt %q: %v", original, err)
			continue
		}

		decrypted, err := service.DecryptString(encrypted)
		if err != nil {
			t.Errorf("Failed to decrypt %q: %v", original, err)
			continue
		}

		if decrypted != original {
			t.Errorf("Decryption result doesn't match original. Got %q, want %q", decrypted, original)
		}
		t.Logf("Decryption successful. Decrypted data: %q", decrypted)
	}
}

func TestWrongKeySize(t *testing.T) {
	// Generate a key of incorrect size
	key, _ := generateRandomKey(30)
	service, _ := NewEncryptionService(key)
	_, err := service.EncryptString("Test data")
	if err == nil {
		t.Error("Expected error when creating EncryptionService with short key, but got none")
	}
}
