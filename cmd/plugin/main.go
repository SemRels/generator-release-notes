// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The plugin-template Authors

package main

import (
	"context"
	"log"
	"os"

	grpcserver "github.com/SemRels/generator-release-notes/internal/grpc"
	semrelplugin "github.com/SemRels/generator-release-notes/internal/plugin"
)

func main() {
	provider := semrelplugin.NewProvider("generator-release-notes")
	server := grpcserver.NewProviderServer(provider)

	if _, err := server.Health(context.Background()); err != nil {
		log.Printf("plugin health check failed: %v", err)
		os.Exit(1)
	}

	log.Printf("%s plugin template is ready", provider.Name())
}
