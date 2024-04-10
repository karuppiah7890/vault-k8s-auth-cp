package main

import (
	"os"

	"github.com/hashicorp/vault/api"
)

type VaultConfig struct {
	Address    string
	Token      string
	CACertPath string
}

func getSourceVaultConfig() VaultConfig {
	// SOURCE_VAULT_ADDR
	// SOURCE_VAULT_TOKEN
	// SOURCE_VAULT_CACERT
	return getVaultConfig("SOURCE_")
}

func getDestinationVaultConfig() VaultConfig {
	// DESTINATION_VAULT_ADDR
	// DESTINATION_VAULT_TOKEN
	// DESTINATION_VAULT_CACERT
	return getVaultConfig("DESTINATION_")
}

func getVaultConfig(envPrefix string) VaultConfig {
	config := VaultConfig{}

	// <envPrefix>VAULT_ADDR
	if address := os.Getenv(envPrefix + api.EnvVaultAddress); address != "" {
		config.Address = address
	}

	// <envPrefix>VAULT_TOKEN
	if token := os.Getenv(envPrefix + api.EnvVaultToken); token != "" {
		config.Token = token
	}

	// <envPrefix>VAULT_CACERT
	if caCertPath := os.Getenv(envPrefix + api.EnvVaultCACert); caCertPath != "" {
		config.CACertPath = caCertPath
	}

	return config
}
