/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	"github.com/XCiber/amplitude-exporter/internal/amplitude"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

func main() {
	initConfig()

	p := amplitude.Projects{}
	err := viper.UnmarshalKey("projects", &p)
	if err != nil {
		log.Fatal("config error: ", err)
	}

	e := amplitude.New(amplitude.SetProjects(&p))
	e.StartScrape()

	prometheus.MustRegister(e)

	s := http.Server{
		Addr:        viper.GetString("listen"),
		ReadTimeout: 2 * time.Second,
	}

	http.Handle("/metrics", promhttp.Handler())

	err = s.ListenAndServe()
	if err != nil {
		log.Fatal("couldn't start server: ", err)
	}
	log.Infof("Beginning to serve on %s", s.Addr)
}

func initConfig() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	viper.SetDefault("listen", ":8080")

	viper.SetEnvPrefix("SRE")
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info("Use config file: ", viper.ConfigFileUsed())
	}

	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
}
