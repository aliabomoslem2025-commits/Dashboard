package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr, MediaGateAddr, CoreServiceToken string
	SynapseBaseURL, SynapseAdminToken string
	SynapseAdminUser, SynapseAdminPassword string
	DBHost, DBPort, DBName, DBUser, DBPassword, RedisAddr string
	RequestTimeout time.Duration
}

func Load() (Config, error) {
	c := Config{
		HTTPAddr: value("CORE_HTTP_ADDR", ":8081"),
		MediaGateAddr: value("MEDIA_GATE_ADDR", ":8082"),
		CoreServiceToken: strings.TrimSpace(os.Getenv("CORE_SERVICE_TOKEN")),
		SynapseBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SYNAPSE_BASE_URL")), "/"),
		SynapseAdminToken: strings.TrimSpace(os.Getenv("SYNAPSE_ADMIN_TOKEN")),
		SynapseAdminUser: strings.TrimSpace(os.Getenv("SYNAPSE_ADMIN_USER")),
		SynapseAdminPassword: os.Getenv("SYNAPSE_ADMIN_PASSWORD"),
		DBHost: value("DB_HOST", "postgres"), DBPort: value("DB_PORT", "5432"),
		DBName: value("DB_DATABASE", "dashboard"), DBUser: value("DB_USERNAME", "dashboard"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		RedisAddr: value("REDIS_ADDR", value("REDIS_HOST", "redis")+":"+value("REDIS_PORT", "6379")),
		RequestTimeout: 8*time.Second,
	}
	credentialPair := c.SynapseAdminUser != "" && c.SynapseAdminPassword != ""
	partialCredentials := (c.SynapseAdminUser == "") != (c.SynapseAdminPassword == "")
	if partialCredentials {
		return c, errors.New("SYNAPSE_ADMIN_USER and SYNAPSE_ADMIN_PASSWORD must be provided together")
	}
	if c.CoreServiceToken == "" || c.SynapseBaseURL == "" || (!credentialPair && c.SynapseAdminToken == "") || c.DBPassword == "" {
		return c, errors.New("CORE_SERVICE_TOKEN, SYNAPSE_BASE_URL, DB_PASSWORD and either SYNAPSE_ADMIN_USER/SYNAPSE_ADMIN_PASSWORD or SYNAPSE_ADMIN_TOKEN are required")
	}
	return c, nil
}

func value(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" { return v }
	return d
}
