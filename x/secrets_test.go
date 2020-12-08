package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCertBase64(t *testing.T) {
	cert := "-----BEGIN CERTIFICATE-----\\nLINE1\\nLINE2-----END CERTIFICATE-----\\n"
	assert.Equal(t, extractCertBase64(cert), "LINE1LINE2")
}

func TestExtractKeyBase64(t *testing.T) {
	cert := "-----BEGIN RSA PRIVATE KEY-----\\nLINE1\\nLINE2-----END RSA PRIVATE KEY-----\\n"
	assert.Equal(t, extractKeyBase64(cert), "LINE1LINE2")
}
