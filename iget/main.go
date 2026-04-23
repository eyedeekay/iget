package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	i "github.com/eyedeekay/iget"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:          "iget [URL]",
	Short:        "i2p terminal HTTP client",
	Long:         `iget is a highly-configurable curl/wget-like client that works exclusively over i2p via the SAM API.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.iget.yaml)")

	// application options
	rootCmd.Flags().String("bridge-host", "127.0.0.1", "host of the SAM bridge")
	rootCmd.Flags().String("bridge-port", "7656", "port of the SAM bridge")
	rootCmd.Flags().StringP("bridge-addr", "p", "", "host:port of the SAM bridge (overrides bridge-host/bridge-port)")
	rootCmd.Flags().String("url", "", "i2p URL to retrieve")

	// debug options
	rootCmd.Flags().Bool("conn-debug", false, "print connection debug info")
	rootCmd.Flags().BoolP("verbose", "v", false, "print additional info about the process")
	rootCmd.Flags().StringP("output", "o", "-", "output path (- for stdout)")

	// i2p options
	rootCmd.Flags().Int("lifespan", 6, "lifespan of an idle i2p destination in minutes")
	rootCmd.Flags().IntP("timeout", "t", 6, "timeout duration in minutes")
	rootCmd.Flags().Int("tunlength", 3, "tunnel length")
	rootCmd.Flags().Int("in-tunnels", 8, "inbound tunnel count")
	rootCmd.Flags().Int("out-tunnels", 8, "outbound tunnel count")
	rootCmd.Flags().Int("in-backups", 3, "inbound backup count")
	rootCmd.Flags().Int("out-backups", 3, "outbound backup count")

	// transport options
	rootCmd.Flags().Bool("disable-keepalives", false, "disable keepalives")
	rootCmd.Flags().Int("idle-conns", 4, "maximum idle connections per host")

	// request options
	rootCmd.Flags().String("method", "GET", "request method")
	rootCmd.Flags().Bool("close", true, "close the request immediately after reading the response")
	rootCmd.Flags().StringArrayP("header", "H", nil, "add a request header in key=value form (repeatable)")

	// eepget / wget compatibility options
	rootCmd.Flags().IntP("line-length", "l", 80, "control line length of output (0 = unlimited)")
	rootCmd.Flags().Int("lineLength", 0, "control line length of output (eepget compat alias)")
	rootCmd.Flags().MarkHidden("lineLength")
	rootCmd.Flags().StringP("etag", "e", "", "set the etag header (not yet implemented)")
	rootCmd.Flags().IntP("mark-size", "m", 0, "show download progress (any value > 0 enables)")
	rootCmd.Flags().IntP("retries", "n", 3, "number of retries")
	rootCmd.Flags().StringP("username", "u", "", "username for SAM authentication")
	rootCmd.Flags().StringP("password", "x", "", "password for SAM authentication")
	rootCmd.Flags().BoolP("continue", "c", true, "resume file from previous download")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/iget")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".iget")
	}
	viper.SetEnvPrefix("IGET")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "warning: error reading config file:", err)
		}
	}
}

func run(cmd *cobra.Command, args []string) error {
	address := viper.GetString("url")
	if len(args) == 1 {
		address = args[0]
	}
	if address == "" {
		return fmt.Errorf("no URL supplied (pass --url or provide as argument)")
	}
	if _, err := url.ParseRequestURI(address); err != nil {
		address = "http://" + address
	}

	verbose := viper.GetBool("verbose")
	if !verbose {
		devNull, err := os.Open(os.DevNull)
		if err == nil {
			syscall.Dup2(int(devNull.Fd()), int(os.Stderr.Fd()))
			devNull.Close()
		}
	}

	samHost := viper.GetString("bridge-host")
	samPort := viper.GetString("bridge-port")
	samAddr := viper.GetString("bridge-addr")

	// Translate eepget-style HTTP proxy addresses (port 4444) to SAM bridge
	if strings.Contains(samAddr, "4444") {
		fmt.Fprintln(os.Stdout, "This application uses the SAM API instead of the http proxy.")
		fmt.Fprintln(os.Stdout, "Please modify your scripts to use the SAM port.")
		return nil
	}

	if samAddr != "" {
		parts := strings.Split(samAddr, ":")
		switch len(parts) {
		case 2:
			samHost = parts[0]
			samPort = parts[1]
		case 1:
			samPort = parts[0]
		}
	}

	lineLen := viper.GetInt("line-length")
	if ll2 := viper.GetInt("lineLength"); ll2 != 0 {
		lineLen = ll2
	}

	headers, _ := cmd.Flags().GetStringArray("header")

	igetClient, err := i.NewIGet(
		i.Lifespan(viper.GetInt("lifespan")),
		i.Timeout(viper.GetInt("timeout")),
		i.Length(viper.GetInt("tunlength")),
		i.Inbound(viper.GetInt("in-tunnels")),
		i.Outbound(viper.GetInt("out-tunnels")),
		i.KeepAlives(viper.GetBool("disable-keepalives")),
		i.Idles(viper.GetInt("idle-conns")),
		i.InboundBackups(viper.GetInt("in-backups")),
		i.OutboundBackups(viper.GetInt("out-backups")),
		i.Debug(viper.GetBool("conn-debug")),
		i.URL(address),
		i.SamHost(samHost),
		i.SamPort(samPort),
		i.Method(viper.GetString("method")),
		i.Output(viper.GetString("output")),
		i.LineLength(lineLen),
		i.Verbose(verbose),
		i.Continue(viper.GetBool("continue")),
		i.Username(viper.GetString("username")),
		i.Password(viper.GetString("password")),
		i.MarkSize(viper.GetInt("mark-size")),
	)
	if err != nil {
		return err
	}

	retries := viper.GetInt("retries")
	for attempt := 0; attempt < retries; attempt++ {
		req, reqErr := igetClient.Request(
			i.Headers(headers),
			i.Close(viper.GetBool("close")),
		)
		if reqErr != nil {
			if attempt == retries-1 {
				return reqErr
			}
			continue
		}
		resp, doErr := igetClient.Do(req)
		if doErr != nil {
			if attempt == retries-1 {
				return doErr
			}
			continue
		}
		igetClient.PrintResponse(resp)
		return nil
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
