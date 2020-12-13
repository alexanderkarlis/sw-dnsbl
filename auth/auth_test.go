package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_JWTCreate(t *testing.T) {
	_, err := CreateJWT("alexanderkarlis", "password", 2)
	require.Equal(t, err, nil, "error should be nil")
}

func Test_JWTValidate(t *testing.T) {
	token, err := CreateJWT("alexanderkarlis", "password", 2)
	require.Equal(t, err, nil, "token creation error should be nil")
	_, err = ValidateToken(token)
	require.Equal(t, err, nil, "validate should be true")
}

func Test_JWTTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in long mode.")
	}
	token, err := CreateJWT("alexanderkarlis", "password", 1)
	require.Equal(t, err, nil, "token creation error should be nil")
	time.Sleep(61 * time.Second)
	_, err = ValidateToken(token)
	require.NotEqual(t, err, nil, "token should have expired")
}
