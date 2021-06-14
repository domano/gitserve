package cmd

import (
	"github.com/domano/gitserve/internal"
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
	RunE:  serveHTTP,
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

func serveHTTP(cmd *cobra.Command, args []string) error {

	url = args[0]
	pk, err := rootCmd.PersistentFlags().GetString("privateKey")
	if err != nil {
		return err
	}
	opts := cloneOpts(url, pk)

	interval, err := rootCmd.PersistentFlags().GetDuration("interval")
	if err != nil {
		return err
	}

	addr, err := rootCmd.PersistentFlags().GetString("address")
	if err != nil {
		return err
	}

	internal.Serve(baseCtx, baseCtxCancel, &opts, interval, addr)
	return nil
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
