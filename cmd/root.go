package cmd

import (
	"github.com/XCiber/amplitude-exporter/internal/amplitude"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	cfgFile = "config"
	verbose = false
	listen  = ":8080"
	timeout = 30
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "amplitude-exporter",
	Short: "Amplitude exporter",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		p := amplitude.Projects{}
		err := viper.UnmarshalKey("projects", &p)
		if err != nil {
			log.Fatal("config error: ", err)
		}

		httpClient := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		e := amplitude.New(amplitude.SetProjects(&p), amplitude.SetHTTPClient(httpClient))

		r := prometheus.NewRegistry()
		r.MustRegister(e)
		handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

		e.StartScrape()

		s := http.Server{
			Addr:        listen,
			ReadTimeout: 5 * time.Second,
		}

		http.Handle("/metrics", handler)

		err = s.ListenAndServe()
		if err != nil {
			log.Fatal("couldn't start server: ", err)
		}
		log.Infof("Beginning to serve on %s", s.Addr)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file name (default is 'config')")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&listen, "listen", "l", ":8080", "listen address")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 30, "timeout for http requests")
}

func initConfig() {

	viper.SetEnvPrefix("SRE")
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName(cfgFile)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/amplitude-exporter/")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err == nil {
		log.Info("Use config file: ", viper.ConfigFileUsed())
	}

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}
