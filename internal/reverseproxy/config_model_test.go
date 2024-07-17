package reverseproxy

import (
	"os"
	"testing"
)

// Test helper to create a temporary certificate file
func createTempCertFile(t *testing.T, content string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "cert_*.pem")
	if err != nil {
		t.Fatalf("Unable to create temp cert file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Unable to write to temp cert file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Unable to close temp cert file: %v", err)
	}

	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

func TestValidateConfig(t *testing.T) {
	validCertContent := "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkq ...\n-----END CERTIFICATE-----"

	// Create valid cert files
	certFile, cleanupCert := createTempCertFile(t, validCertContent)
	defer cleanupCert()

	keyFile, cleanupKey := createTempCertFile(t, validCertContent)
	defer cleanupKey()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Routes: []Route{
					{
						Name:       "route1",
						ListenHost: "localhost",
						ListenPort: 8080,
						Protocol:   "http",
						Pattern:    "/api",
					},
					{
						Name:       "route2",
						ListenHost: "localhost",
						ListenPort: 8443,
						Protocol:   "https",
						Pattern:    "/secure",
						CertFile:   certFile,
						KeyFile:    keyFile,
						Target: Target{
							Name:     "target1",
							Protocol: "https",
							Host:     "example.com",
							Port:     443,
							CertFile: certFile,
							KeyFile:  keyFile,
							CaCert:   certFile,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid listen port",
			config: Config{
				Routes: []Route{
					{
						Name:       "route1",
						ListenHost: "localhost",
						ListenPort: 70000, // invalid port
						Protocol:   "http",
						Pattern:    "/api",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid protocol",
			config: Config{
				Routes: []Route{
					{
						Name:       "route1",
						ListenHost: "localhost",
						ListenPort: 8080,
						Protocol:   "ftp", // invalid protocol
						Pattern:    "/api",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing certificate file for https",
			config: Config{
				Routes: []Route{
					{
						Name:       "route1",
						ListenHost: "localhost",
						ListenPort: 8443,
						Protocol:   "https",
						Pattern:    "/secure",
						CertFile:   "nonexistent.pem", // Cert file doesn't exist
						KeyFile:    "nonexistent.pem",
						Target: Target{
							Name:     "target1",
							Protocol: "https",
							Host:     "example.com",
							Port:     443,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
