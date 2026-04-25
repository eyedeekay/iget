// Package main is the entry point for the iget command-line binary.
// It wires cobra flags and viper configuration to the iget library,
// providing a curl/wget-like interface for downloading resources over I2P.
package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	i "github.com/go-i2p/iget"
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
	rootCmd.Flags().StringP("etag", "e", "", "set the If-None-Match request header for conditional GETs")
	rootCmd.Flags().IntP("mark-size", "m", 0, "show download progress (any value > 0 enables)")
	rootCmd.Flags().IntP("retries", "n", 3, "number of retries")
	rootCmd.Flags().StringP("data", "d", "", "request body for POST/PUT")
	rootCmd.Flags().StringP("username", "u", "", "username for SAM authentication")
	rootCmd.Flags().StringP("password", "x", "", "password for SAM authentication")
	rootCmd.Flags().BoolP("continue", "c", true, "resume file from previous download")
	rootCmd.Flags().String("to-port", "", "SAM virtual destination port")
	rootCmd.Flags().String("from-port", "", "SAM virtual source port")
	rootCmd.Flags().String("session-name", "", "SAM session name (default: unique per invocation; set to reuse a persistent I2P identity)")
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

// parseSAMBridgeAddr parses a SAM bridge address in host:port form, correctly
// handling IPv6 bracket notation (e.g. "[::1]:7656"). It returns an error when
// the address is not a valid host:port pair.
func parseSAMBridgeAddr(addr string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return "", "", fmt.Errorf("invalid --bridge-addr %q: %w", addr, err)
	}
	return host, port, nil
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

	samHost := viper.GetString("bridge-host")
	samPort := viper.GetString("bridge-port")
	samAddr := viper.GetString("bridge-addr")

	// Translate eepget-style HTTP proxy addresses (port 4444) to SAM bridge
	if samAddr != "" {
		host, port, err := parseSAMBridgeAddr(samAddr)
		if err != nil {
			return err
		}
		if port == "4444" {
			fmt.Fprintln(os.Stdout, "This application uses the SAM API instead of the http proxy.")
			fmt.Fprintln(os.Stdout, "Please modify your scripts to use the SAM port.")
			return nil
		}
		samHost = host
		samPort = port
	}

	lineLen := viper.GetInt("line-length")
	if ll2 := viper.GetInt("lineLength"); ll2 != 0 {
		lineLen = ll2
	}

	headers, _ := cmd.Flags().GetStringArray("header")
	if etag := viper.GetString("etag"); etag != "" {
		headers = append(headers, "If-None-Match="+etag)
	}

	igetClient, err := i.NewIGet(
		i.Lifespan(viper.GetInt("lifespan")),
		i.Timeout(viper.GetInt("timeout")),
		i.Length(viper.GetInt("tunlength")),
		i.Inbound(viper.GetInt("in-tunnels")),
		i.Outbound(viper.GetInt("out-tunnels")),
		i.DisableKeepAlives(viper.GetBool("disable-keepalives")),
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
		i.Body(viper.GetString("data")),
		i.ToPort(viper.GetString("to-port")),
		i.FromPort(viper.GetString("from-port")),
		i.SessionName(viper.GetString("session-name")),
	)
	if err != nil {
		return err
	}
	defer igetClient.Close()

	// Install a signal handler so that Ctrl+C (SIGINT) and SIGTERM trigger
	// SAM session cleanup. Without this, Go's default handler calls os.Exit
	// directly, bypassing deferred functions and leaving the SAM session open
	// in the I2P router until its idle timeout expires.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		igetClient.Close() //nolint:errcheck
		os.Exit(1)
	}()

	retries := viper.GetInt("retries")
	if retries < 1 {
		retries = 1
	}
	for attempt := 0; attempt < retries; attempt++ {
		if attempt > 0 {
			// Back-off between retries so that transient I2P tunnel failures have
			// time to recover. The delay grows linearly with the attempt number so
			// the first retry is quick (5 s) and later ones are progressively slower.
			time.Sleep(time.Duration(attempt) * 5 * time.Second)
			// Recreate the SAM session so that a stale or broken session from the
			// previous attempt does not cause all retries to fail immediately.
			if resetErr := igetClient.Reset(); resetErr != nil {
				if attempt == retries-1 {
					return resetErr
				}
				continue
			}
		}
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
