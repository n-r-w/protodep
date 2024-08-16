package cmd

import (
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/n-r-w/protodep/internal/logger"
	"github.com/n-r-w/protodep/internal/resolver"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Populate .proto vendors existing protodep.toml",
	RunE: func(cmd *cobra.Command, _ []string) error {
		isCleanupCache, err := cmd.Flags().GetBool("cleanup")
		if err != nil {
			return err
		}
		logger.Info("cleanup cache = %t", isCleanupCache)

		identityFile, err := cmd.Flags().GetString("identity-file")
		if err != nil {
			return err
		}
		logger.Info("identity file = %s", identityFile)

		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		if password != "" {
			logger.Info("password = %s", strings.Repeat("x", len(password))) // Do not display the password.
		}

		useHTTPS, err := cmd.Flags().GetBool("use-https")
		if err != nil {
			return err
		}
		logger.Info("use https = %t", useHTTPS)

		useNetrc, err := cmd.Flags().GetBool("use-netrc")
		if err != nil {
			return err
		}
		logger.Info("use netrc = %t", useNetrc)

		basicAuthUsername, err := cmd.Flags().GetString("basic-auth-username")
		if err != nil {
			return err
		}
		if basicAuthUsername != "" {
			logger.Info("https basic auth username = %s", basicAuthUsername)
		}

		basicAuthPassword, err := cmd.Flags().GetString("basic-auth-password")
		if err != nil {
			return err
		}
		if basicAuthPassword != "" {
			logger.Info("https basic auth password = %s", strings.Repeat("x", len(basicAuthPassword))) // Do not display the password.
		}

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		homeDir, err := homedir.Dir()
		if err != nil {
			return err
		}

		conf := resolver.Config{
			UseHttps:          useHTTPS,
			UseNetrc:          useNetrc,
			HomeDir:           homeDir,
			TargetDir:         pwd,
			OutputDir:         pwd,
			BasicAuthUsername: basicAuthUsername,
			BasicAuthPassword: basicAuthPassword,
			IdentityFile:      identityFile,
			IdentityPassword:  password,
		}

		httpsProvider, err := conf.GetHttpsAuthProvider()
		if err != nil {
			return err
		}

		sshProvider, err := conf.GetSshAuthProvider()
		if err != nil {
			return err
		}

		updateService, err := resolver.New(&conf, httpsProvider, sshProvider)
		if err != nil {
			return err
		}

		return updateService.Resolve(isCleanupCache)
	},
}

func initDepCmd() {
	upCmd.PersistentFlags().StringP("identity-file", "i", "", "set the identity file for SSH")
	upCmd.PersistentFlags().StringP("password", "p", "", "set the password for SSH")
	upCmd.PersistentFlags().BoolP("cleanup", "c", false, "cleanup cache before exec.")
	upCmd.PersistentFlags().BoolP("use-https", "u", false, "use HTTPS to get dependencies.")
	upCmd.PersistentFlags().BoolP("use-netrc", "n", true, "use netrc file for authentication")
	upCmd.PersistentFlags().StringP("basic-auth-username", "", "", "set the username with Basic Auth via HTTPS")
	upCmd.PersistentFlags().StringP("basic-auth-password", "", "", "set the password or personal access token(when enabled 2FA) with Basic Auth via HTTPS")
}
