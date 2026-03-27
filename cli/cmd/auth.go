package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/dawillygene/my-prompt-repository/internal/config"
	"github.com/spf13/cobra"
)

var (
	registerName     string
	registerEmail    string
	registerPassword string
	loginEmail       string
	loginPassword    string
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user account",
	Long:  `Create a new user account and receive an authentication token.`,
	Example: `  prompt register --name "John Doe" --email john@example.com --password secret123
  prompt register --json --name "Dev User" --email dev@test.com --password pass`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		cfg := getConfig()

		response, err := client.Request("POST", "/api/register", map[string]any{
			"name":     registerName,
			"email":    registerEmail,
			"password": registerPassword,
		}, false)
		if err != nil {
			return err
		}

		token, _ := response["token"].(string)
		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			return err
		}
		setConfig(cfg)

		return prettyPrint(response)
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your account",
	Long:  `Authenticate with your email and password to receive an access token.`,
	Example: `  prompt login --email john@example.com --password secret123
  prompt login --json --email dev@test.com --password pass`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		cfg := getConfig()

		response, err := client.Request("POST", "/api/login", map[string]any{
			"email":    loginEmail,
			"password": loginPassword,
		}, false)
		if err != nil {
			return err
		}

		token, _ := response["token"].(string)
		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			return err
		}
		setConfig(cfg)

		return prettyPrint(response)
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of your account",
	Long:  `Invalidate your current authentication token and log out.`,
	Example: `  prompt logout
  prompt logout --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		cfg := getConfig()

		if _, err := client.Request("POST", "/api/logout", map[string]any{}, true); err != nil {
			return err
		}

		cfg.Token = ""
		if err := config.Save(cfg); err != nil {
			return err
		}
		setConfig(cfg)

		if isJSONMode() {
			return prettyPrint(map[string]any{
				"message": "Logged out.",
				"data":    map[string]any{},
			})
		}

		fmt.Println("Logged out.")
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  `Show details about the currently authenticated user.`,
	Example: `  prompt whoami
  prompt whoami --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		response, err := client.Request("GET", "/api/me", nil, true)
		if err != nil {
			return err
		}
		return prettyPrint(response)
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)

	// Register flags
	registerCmd.Flags().StringVar(&registerName, "name", "", "your full name")
	registerCmd.Flags().StringVar(&registerEmail, "email", "", "your email address")
	registerCmd.Flags().StringVar(&registerPassword, "password", "", "your password")
	registerCmd.MarkFlagRequired("name")
	registerCmd.MarkFlagRequired("email")
	registerCmd.MarkFlagRequired("password")

	// Login flags
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "your email address")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "your password")
	loginCmd.MarkFlagRequired("email")
	loginCmd.MarkFlagRequired("password")
}

// Helper function to print JSON
func prettyPrint(value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(content))
	return nil
}
