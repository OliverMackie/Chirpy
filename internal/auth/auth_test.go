package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMakeJWT(t *testing.T) {
	userId := uuid.New()
	tokenSecret := "12345"
	ExpiresIn := time.Hour
	_, err := MakeJWT(userId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestValidateJWT(t *testing.T) {
	initId := uuid.New()
	userId := initId.String()
	tokenSecret := "12345"
	ExpiresIn := time.Hour
	token, err := MakeJWT(initId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	id, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if id != initId {
		t.Fatalf("expected id %s, got %v", userId, id)
	}
}

func TestValidateJWTWrongKey(t *testing.T) {
	initId := uuid.New()

	tokenSecret := "12345"
	wrong := "123456"
	ExpiresIn := time.Hour
	token, err := MakeJWT(initId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	err2 := fmt.Sprintf("token signature is invalid: signature is invalid")
	_, err = ValidateJWT(token, wrong)
	assert.EqualErrorf(t, err, err2, "Error should be: %v, got: %v", err2, err)

}

func TestValidateJWTExpired(t *testing.T) {
	initId := uuid.New()
	tokenSecret := "12345"
	ExpiresIn := time.Second
	token, err := MakeJWT(initId, tokenSecret, ExpiresIn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	time.Sleep(2 * time.Second)
	err2 := fmt.Sprintf("token has invalid claims: token is expired")
	_, err = ValidateJWT(token, tokenSecret)
	assert.EqualErrorf(t, err, err2, "Error should be: %v, got: %v", err2, err)

}

func TestGetBearerToken(t *testing.T) {
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer 12345"}
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	assert.Equal(t, "12345", token)
}

func TestGetBearerTokenMissing(t *testing.T) {
	headers := make(map[string][]string)
	_, err := GetBearerToken(headers)
	err2 := fmt.Sprintf("authorization header is missing")
	assert.EqualErrorf(t, err, err2, "Error should be: %v, got: %v", err2, err)
}

func TestMakeRefreshToken(t *testing.T) {
	token := MakeRefreshToken()
	if len(token) != 64 {
		t.Fatalf("expected token length 64, got %d", len(token))
	}
}

func TestGetAPIKey(t *testing.T) {
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"ApiKey 12345"}
	token, err := GetAPIKey(headers)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	assert.Equal(t, "12345", token)
}
