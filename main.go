package main

// This is a CLI tool to copy Kubernetes Auth Config and Roles from one Vault to another Vault

// We will be primarily using Vault Official Golang API to interact with Vault Server.
// Vault Official Golang API Docs are here -
// https://pkg.go.dev/github.com/hashicorp/vault/api
// This Official Golang API is part of the Vault source code itself :D

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

var usage = `usage: vault-k8s-auth-cp <source-k8s-auth-mount-path> <destination-k8s-auth-mount-path>
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s", usage)
		os.Exit(0)
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
	}

	// Get Config for Source Vault
	sourceConfig := getSourceVaultConfig()

	// Create a new client to the source vault
	sourceDefaultConfig := api.DefaultConfig()
	sourceDefaultConfig.Address = sourceConfig.Address

	if sourceConfig.CACertPath != "" {
		sourceDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: sourceConfig.CACertPath})
	}
	sourceClient, err := api.NewClient(sourceDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating source vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the source vault client
	sourceClient.SetToken(sourceConfig.Token)

	// Get Config for Destination Vault
	destinationConfig := getDestinationVaultConfig()

	// Create a new client to the destination vault
	destinationDefaultConfig := api.DefaultConfig()
	destinationDefaultConfig.Address = destinationConfig.Address
	if destinationConfig.CACertPath != "" {
		destinationDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: destinationConfig.CACertPath})
	}
	destinationClient, err := api.NewClient(destinationDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating destination vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the destination vault client
	destinationClient.SetToken(destinationConfig.Token)

	// Get the path to the secrets in the source vault.
	// Get the path to the secrets in the destination vault
	var sourceMountPath, destinationMountPath string

	sourceMountPath = flag.Args()[0]
	destinationMountPath = flag.Args()[1]

	_ = destinationMountPath

	// Get the Kubernetes Auth Config and Roles from the source vault and keep it in memory.
	// Create the Kubernetes Auth Config and Roles in the destination vault.
	// If the destination vault already has the same mounth path, then it should be overwritten.
	// OR
	// Directly copy the Kubernetes Auth Config and Roles from the source Vault to the destination Vault
	// without keeping it in memory.
	// If the destination vault already has the same mounth path, then it should be overwritten.

	// For now, assume that the destination vault does already have the same mount path with/without some existing data,
	// meaning that Kubernetes auth has been enabled at that path in the destination vault.
	// This tool will NOT enable Kubernetes Auth at the destination path for now.

	// TODO: Enable Kubernetest Auth at the destination path if it is not already enabled.

	// Get the Kubernetes Auth Config from the source vault
	// NOTE: This does NOT read the Token Reviewer JWT from the source vault
	sourceK8sAuthConfig, err := sourceClient.Logical().Read("auth/" + sourceMountPath + "/config")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading k8s auth config from source vault: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSource Kubernetes Auth Config: %+v\n", sourceK8sAuthConfig.Data)

	// Create the Kubernetes Auth Config in the destination vault
	// TODO: Support writing Token Reviewer JWT along with this, as it does NOT get read from the source vault
	destinationK8sAuthConfig, err := destinationClient.Logical().Write("auth/"+destinationMountPath+"/config", sourceK8sAuthConfig.Data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing k8s auth config to destination vault: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDestination Kubernetes Auth Config: %+v\n", destinationK8sAuthConfig)

	// Get the Kubernetes Auth Roles from the source vault
	sourceK8sAuthRolesInfo, err := sourceClient.Logical().List("auth/" + sourceMountPath + "/role")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing the k8s auth roles from source vault: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSource Kubernetes Auth Roles: %+v\n", sourceK8sAuthRolesInfo.Data)

	sourceK8sAuthRoles := sourceK8sAuthRolesInfo.Data["keys"].([]interface{})
	for _, role := range sourceK8sAuthRoles {
		// Get the Kubernetes Auth Role from the source vault
		roleName := role.(string)
		sourceK8sAuthRole, err := sourceClient.Logical().Read("auth/" + sourceMountPath + "/role/" + roleName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading k8s auth role from source vault: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSource Kubernetes Auth Role: %+v\n", sourceK8sAuthRole.Data)

		// Create the Kubernetes Auth Role in the destination vault
		destinationK8sAuthRole, err := destinationClient.Logical().Write("auth/"+destinationMountPath+"/role/"+roleName, sourceK8sAuthRole.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing k8s auth role to destination vault: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nDestination Kubernetes Auth Role: %+v\n", destinationK8sAuthRole)
	}

}
