package cmd

import (
	"github.com/domano/gitserve/internal"
	"github.com/go-git/go-git/v5"
	"log"
	"regexp"

	"github.com/spf13/cobra"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "Serves the repository via http.",
	Long:  `Serves the repository via the golang http fileserver.`,
	Args:  cobra.RangeArgs(1, 1),
	Run:   serveHTTP,
}

func init() {
	rootCmd.AddCommand(httpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func serveHTTP(cmd *cobra.Command, args []string) {

	url = args[0]

	opts := cloneOpts(url, privateKey)

	internal.Serve(baseCtx, baseCtxCancel, &opts, updateInterval, addr)
}

func cloneOpts(url, privateKey string) git.CloneOptions {
	url = parseUrl(url)

	pk := getPublicKey(privateKey)

	opts := git.CloneOptions{
		URL:               url,
		Auth:              pk,
		RemoteName:        "",
		ReferenceName:     "",
		SingleBranch:      false,
		NoCheckout:        false,
		Depth:             0,
		RecurseSubmodules: 0,
		Progress:          nil,
		Tags:              0,
		InsecureSkipTLS:   false,
		CABundle:          nil,
	}
	return opts
}

func parseUrl(url string) string {
	// if we do not have something like a protocol specifier in front or it does not look like a ssh-url we append https:// as a default
	match, err := regexp.Match(".*(://|@.*:).*", []byte(url))
	if err != nil {
		log.Fatalln(err)
	}
	if !match {
		url = "https://" + url
	}
	return url
}
