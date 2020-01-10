package cmd

import (
	"strings"
	"github.com/go-shiori/shiori/internal/ldap"
	"github.com/go-shiori/shiori/internal/webserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve web interface for managing bookmarks",
		Long: "Run a simple and performant web server which " +
			"serves the site for managing bookmarks. If --port " +
			"flag is not used, it will use port 8080 by default.",
		Run: serveHandler,
	}

	cmd.Flags().IntP("port", "p", 8080, "Port used by the server")
	cmd.Flags().StringP("address", "a", "", "Address the server listens to")
	cmd.Flags().StringP("webroot", "r", "/", "Root path that used by server")
	cmd.Flags().StringP("ldap", "", "", "Path to config file for connecting with LDAP server")

	return cmd
}

func serveHandler(cmd *cobra.Command, args []string) {
	// Get flags value
	port, _ := cmd.Flags().GetInt("port")
	address, _ := cmd.Flags().GetString("address")
	rootPath, _ := cmd.Flags().GetString("webroot")

	// Validate root path
	if rootPath == "" {
		rootPath = "/"
	}

	if !strings.HasPrefix(rootPath, "/") {
		rootPath = "/" + rootPath
	}

	if !strings.HasSuffix(rootPath, "/") {
		rootPath += "/"
	}

	ldapConfigPath, _ := cmd.Flags().GetString("ldap")

	options := webserver.Options{
		DB:         db,
		DataDir:    dataDir,
		Address:    address,
		Port:       port,
		RootPath:      rootPath,
		LDAPClient: nil,
	}

	if ldapConfigPath != "" {
		cfg, err := ldap.ParseConfigFile(ldapConfigPath)
		if err != nil {
			logrus.Fatalf("Failed to open LDAP config: %v\n", err)
		}

		options.LDAPClient, err = ldap.NewClient(cfg)
		if err != nil {
			logrus.Fatalf("Failed to create LDAP client: %v\n", err)
		}
		defer options.LDAPClient.Close()
	}

  // Start server
	err := webserver.ServeApp(options)
	if err != nil {
		logrus.Fatalf("Server error: %v\n", err)
	}
}
