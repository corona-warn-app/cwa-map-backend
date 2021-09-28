package main

import (
	"com.t-systems-mms.cwa/external/geocoding"
	"com.t-systems-mms.cwa/services"
	"errors"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	Logging        LogConfig
	Database       DatabaseConfig
	Google         geocoding.GoogleGeocoderConfig
	Authentication AuthenticationConfig
	BugReports     services.BugReportConfig
	Email          services.EmailConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	IdlePoolSize    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type ServerConfig struct {
	Listen string
}

type AuthenticationConfig struct {
	JwksUrl string
	KeyId   string
	KeyAlg  string
}

type LogConfig struct {
	Level  string
	LogSQL bool
}

var appConfig = &Config{}

// LoadConfig loads the application config.
// If the functions an error, the application should panic (or just exit),
// cause it means that some critical configuraion values could not be loaded
func LoadConfig() error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	if err := kubernetesLogin(client); err != nil {
		return err
	}
	logicalClient := client.Logical()
	backend := os.Getenv("CWA_MAP_VAULT_BACKEND")
	if backend == "" {
		return errors.New("missing CWA_MAP_VAULT_BACKEND")
	}

	appConfig.Server.Listen = getEnv("CWA_MAP_SERVER_LISTEN", ":9090")
	appConfig.Logging.Level = getEnv("CWA_MAP_LOG_LEVEL", "info")
	appConfig.Logging.LogSQL, err = strconv.ParseBool(getEnv("CWA_MAP_LOG_SQL", "false"))
	if err != nil {
		return err
	}

	// Database
	if err := readStringSecret(logicalClient, backend+"/data/database", "host",
		&appConfig.Database.Host); err != nil {
		return err
	}
	if err := readIntSecret(logicalClient, backend+"/data/database", "port",
		&appConfig.Database.Port); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/database", "database",
		&appConfig.Database.Database); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/database", "user",
		&appConfig.Database.User); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/database", "password",
		&appConfig.Database.Password); err != nil {
		return err
	}
	if err := readIntSecret(logicalClient, backend+"/data/database", "max-open-conns",
		&appConfig.Database.MaxOpenConns); err != nil {
		appConfig.Database.MaxOpenConns = 100
	}
	if err := readIntSecret(logicalClient, backend+"/data/database", "conn-max-lifetime",
		&appConfig.Database.ConnMaxLifetime); err != nil {
		appConfig.Database.ConnMaxLifetime = 60
	}
	if err := readIntSecret(logicalClient, backend+"/data/database", "idle-pool-size",
		&appConfig.Database.IdlePoolSize); err != nil {
		appConfig.Database.IdlePoolSize = 10
	}

	// E-Mail
	if err := readStringSecret(logicalClient, backend+"/data/email", "smtp-host",
		&appConfig.Email.SmtpHost); err != nil {
		return err
	}
	if err := readIntSecret(logicalClient, backend+"/data/email", "smtp-port",
		&appConfig.Email.SmtpPort); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/email", "smtp-user",
		&appConfig.Email.SmtpUser); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/email", "smtp-password",
		&appConfig.Email.SmtpPassword); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/email", "from",
		&appConfig.Email.From); err != nil {
		return err
	}

	// Bug reports
	if err := readIntSecret(logicalClient, backend+"/data/reports", "interval",
		&appConfig.BugReports.Interval); err != nil {
		appConfig.BugReports.Interval = 24 * 60
	}

	// Authentication
	if err := readStringSecret(logicalClient, backend+"/data/authentication", "jwks-url",
		&appConfig.Authentication.JwksUrl); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/authentication", "jwks-key-id",
		&appConfig.Authentication.KeyId); err != nil {
		return err
	}
	if err := readStringSecret(logicalClient, backend+"/data/authentication", "key-alg",
		&appConfig.Authentication.KeyAlg); err != nil {
		return err
	}

	if err := readStringSecret(logicalClient, backend+"/data/google-maps", "api-key",
		&appConfig.Google.ApiKey); err != nil {
		return err
	}

	return nil
}

func kubernetesLogin(client *api.Client) error {
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		// if we have a token, use this, but warn
		logrus.Warn("VAULT_TOKEN found in environment, do not use this in production")
		client.SetToken(token)
		return nil
	}

	jwtContent, err := os.ReadFile(getEnv("CWA_VAULT_TOKENFILE", ""))
	if err != nil {
		return err
	}

	if auth, err := client.Logical().Write("auth/kubernetes/login", map[string]interface{}{
		"role": os.Getenv("VAULT_ROLE"),
		"jwt":  string(jwtContent),
	}); err == nil {
		client.SetToken(auth.Auth.ClientToken)
		return nil
	} else {
		return err
	}
}

// getEnv reads the value of the given environment variable or returns a default value, if the variable does not exist
func getEnv(name, defaultValue string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return defaultValue
}

// readStringSecret reads the string secret from the given Vault connection
func readStringSecret(logical *api.Logical, path, name string, target *string) error {
	if secret, err := logical.Read(path); err != nil {
		return err
	} else {
		if secret == nil || secret.Data == nil {
			return errors.New(fmt.Sprintf("path %s does not exist", path))
		}

		secretData := secret.Data["data"].(map[string]interface{})
		if _, exists := secretData[name]; !exists {
			return errors.New(fmt.Sprintf("secret %s/%s does not exist in path", path, name))
		}

		secretValue := secretData[name]
		if secretString, ok := secretValue.(string); ok {
			*target = secretString
			return nil
		} else {
			return errors.New(fmt.Sprintf("secret %s/%s has invalid type (%v)",
				path, name, reflect.TypeOf(secretValue)))
		}
	}
}

func readIntSecret(logical *api.Logical, path, name string, target *int) error {
	var stringValue string
	if err := readStringSecret(logical, path, name, &stringValue); err != nil {
		return err
	} else {
		if intValue, err := strconv.Atoi(stringValue); err != nil {
			return err
		} else {
			*target = intValue
			return nil
		}
	}
}
