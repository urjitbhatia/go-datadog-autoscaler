package main

import (
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/urjitbhatia/go-datadog-autoscaler/scaler"
	"github.com/zorkian/go-datadog-api"
)

func main() {

	config := getConfig()
	cl := getDatadogClient(config)
	ddMetrics, ok := config.Get("ddMetrics").([]interface{})
	if !ok {
		log.Fatalln("Error parsing ddMetrics. Lint yaml")
	}

	for _, ddMetric := range ddMetrics {
		metric := &scaler.Metric{}
		mapstructure.Decode(ddMetric, metric)
		scaler.ProcessMetric(*metric, cl)
	}
}

func getDatadogClient(config *viper.Viper) *datadog.Client {

	log.Println("Connecting to Datadog")
	ddApiKey := config.GetString("ddApiKey")
	ddAppKey := config.GetString("ddAppKey")

	if len(ddApiKey) == 0 || len(ddAppKey) == 0 {
		log.Fatalln("Datadog config missing api key or app key")
	}
	client := datadog.NewClient(ddApiKey, ddAppKey)
	return client
}

func getConfig() *viper.Viper {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath(".")

	err := config.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config. Looking for `config.yaml` in the current dir. Err:{%+v}", err)
	}
	return config
}
