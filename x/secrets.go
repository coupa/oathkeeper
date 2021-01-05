package x

import (
	"fmt"
	"os"
	"strings"

	conf "github.com/coupa/foundation-go/config"
	log "github.com/sirupsen/logrus"
)

const (
	AppCertSecretEnvName = "AWS_APP_CERTS_SECRET_NAME"
	appSSLCert           = "app_ssl_certificate"
	appSSLKey            = "app_ssl_certificate_key"

	SecretManagerEnvName = "AWSSM_NAME"
	AWSRegionEnvName     = "AWS_REGION"
	proxyCertEnvName     = "SERVE_PROXY_TLS_CERT_BASE64"
	proxyKeyEnvName      = "SERVE_PROXY_TLS_KEY_BASE64"
	apiCertEnvName       = "SERVE_API_TLS_CERT_BASE64"
	apiKeyEnvName        = "SERVE_API_TLS_KEY_BASE64"
)

func GetMainSecrets() (map[string]string, error) {
	smName := os.Getenv(SecretManagerEnvName)
	if smName == "" {
		return nil, nil
	}
	if os.Getenv(AWSRegionEnvName) == "" {
		//Default region to us-east-1
		if err := os.Setenv(AWSRegionEnvName, "us-east-1"); err != nil {
			return nil, err
		}
	}
	data, _, err := conf.GetSecrets(smName)
	return data, err
}

//SetupAppCerts gets a cert and a key from secrets manager. It extracts the base64
//portion in the cert and the key and sets them to the corresponding env variables.
func SetupAppCerts(appcertsSecretName string) error {
	if appcertsSecretName == "" {
		return nil
	}
	secrets, _, err := conf.GetSecrets(appcertsSecretName)
	if err != nil {
		return fmt.Errorf("Error getting app certs from Secrets Manager: %v", err)
	}

	if os.Getenv(proxyCertEnvName) == "" {
		pem := secrets[appSSLCert]
		if pem == "" {
			return fmt.Errorf("App certificate (%s) on Secrets Manager (%s) not found", appSSLCert, appcertsSecretName)
		}
		cert := extractCertBase64(pem)
		if err = os.Setenv(proxyCertEnvName, cert); err != nil {
			return fmt.Errorf("Error setting %s with cert from secrets manager: %v", proxyCertEnvName, err)
		}
		if err = os.Setenv(apiCertEnvName, cert); err != nil {
			return fmt.Errorf("Error setting %s with cert from secrets manager: %v", apiCertEnvName, err)
		}
		log.Infof("Successfully set TLS cert from Secrets Manager")
	}
	if os.Getenv(proxyKeyEnvName) == "" {
		key := secrets[appSSLKey]
		if key == "" {
			return fmt.Errorf("App key (%s) on Secrets Manager (%s) not found", appSSLKey, appcertsSecretName)
		}
		privKey := extractKeyBase64(key)
		if err = os.Setenv(proxyKeyEnvName, privKey); err != nil {
			return fmt.Errorf("Error setting %s with TLS key from secrets manager: %v", proxyKeyEnvName, err)
		}
		if err = os.Setenv(apiKeyEnvName, privKey); err != nil {
			return fmt.Errorf("Error setting %s with TLS key from secrets manager: %v", apiKeyEnvName, err)
		}
		log.Infof("Successfully set TLS key from Secrets Manager")
	}
	return nil
}

func extractCertBase64(pem string) string {
	pem = strings.TrimPrefix(pem, "-----BEGIN CERTIFICATE-----\\n")
	pem = strings.TrimSuffix(pem, "-----END CERTIFICATE-----\\n")
	return strings.ReplaceAll(pem, "\\n", "")
}

func extractKeyBase64(pem string) string {
	pem = strings.TrimPrefix(pem, "-----BEGIN RSA PRIVATE KEY-----\\n")
	pem = strings.TrimSuffix(pem, "-----END RSA PRIVATE KEY-----\\n")
	return strings.ReplaceAll(pem, "\\n", "")
}
