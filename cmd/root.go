package cmd

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var privateKey string
var addr string
var userName, password string
var updateInterval time.Duration
var url string

// TODO: make this work for all subcommands
var baseCtx, baseCtxCancel = context.WithCancel(context.Background())

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitserve",
	Short: "Serve any git repository from memory via http.",
	Long:  `Serve any git repository from memory via http whilst keeping it up to date.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&privateKey, "privateKey", "k", "~/.ssh/id_rsa", "For SSH cloning, fetching and pulling you can pass a private key.")
	rootCmd.PersistentFlags().StringVarP(&addr, "address", "a", ":8080", "Address to use for the server.")
	rootCmd.PersistentFlags().StringVarP(&userName, "user", "u", "", "Username, if required.")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Password, if required. PASSING THIS AS A FLAG WILL SHOW THE PASSWORD IN YOUR HISTORY.")
	rootCmd.PersistentFlags().DurationVarP(&updateInterval, "interval", "i", 5*time.Minute, "Interval that determines how often we check and pull in changes from git.")

	viper.BindPFlags(rootCmd.Flags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".gitserve" (without extension).
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigName(".gitserve")

	viper.SetEnvPrefix("gitserve")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to GITSERVE_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", "gitserve", envVarSuffix))
		}

		f.
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
func getPublicKey(privateKey string) *ssh.PublicKeys {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if privateKey == "~" {
		// In case of "~", which won't be caught by the prefix case
		privateKey = dir
	}
	if strings.HasPrefix(privateKey, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		privateKey = filepath.Join(dir, privateKey[2:])
	}

	publicKeys, err := checkPassword(ssh.NewPublicKeysFromFile("git", privateKey, "pw"))
	if err != nil {
		log.Fatal(err)
	}
	return publicKeys
}

func checkPassword(publicKeys *ssh.PublicKeys, err error) (*ssh.PublicKeys, error) {
	if errors.Is(err, x509.IncorrectPasswordError) || (err != nil && strings.Contains(err.Error(), "password")) { // hacky catch-all check for passwords since not all possible password errors are properly typed
		pw, err := promptPassword()
		if err != nil {
			return nil, err
		}
		return checkPassword(ssh.NewPublicKeysFromFile("git", privateKey, pw))
	}
	return publicKeys, err
}

func promptPassword() (string, error) {
	fmt.Println("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(password), nil
}
