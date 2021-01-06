package x

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessAndEncode(t *testing.T) {
	cert := "-----BEGIN CERTIFICATE-----\\nLINE1\\nLINE2-----END CERTIFICATE-----\\n"
	expected := "-----BEGIN CERTIFICATE-----\nLINE1\nLINE2-----END CERTIFICATE-----\n"

	assert.Equal(t, processAndEncode(cert), base64.StdEncoding.EncodeToString([]byte(expected)))

	key := "-----BEGIN RSA PRIVATE KEY-----\\nLINE1\\nLINE2-----END RSA PRIVATE KEY-----\\n"
	expected = "-----BEGIN RSA PRIVATE KEY-----\nLINE1\nLINE2-----END RSA PRIVATE KEY-----\n"

	assert.Equal(t, processAndEncode(key), base64.StdEncoding.EncodeToString([]byte(expected)))
}
