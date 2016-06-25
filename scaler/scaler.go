package scaler

import (
	htime "github.com/urjitbhatia/gohumantime"
	"github.com/zorkian/go-datadog-api"
	"log"
	"math"
	"time"
)

const (
	lastTransform  = "last"
	avgTransform   = "avg"
	minTransform   = "min"
	maxTransform   = "max"
	sumTransform   = "sum"
	countTransform = "count"
)

type Scale struct {
	Count     int64
	Threshold float64
	Cooldown  bool
	GroupName string
}

type Metric struct {
	Name      string
	Query     string
	Period    string
	Transform string
	ScaleUp   *Scale
	ScaleDown *Scale
}

func ProcessMetric(metric Metric, client *datadog.Client) {

	if len(metric.Query) == 0 {
		log.Fatalln("Cannot read query for metric: %s", metric.Name)
	}
	var durationFromEnd = time.Minute * 5 // Default is 5 minute
	period := metric.Period
	if len(period) > 0 {
		millisFromEnd, err := htime.ToMilliseconds(period)
		if err == nil {
			durationFromEnd = time.Duration(millisFromEnd) * time.Millisecond
		}
	} else {
		log.Printf("Using a default period of 5 minutes for metric: %s", metric.Name)
	}

	end := time.Now()
	start := end.Add(-1 * durationFromEnd)
	log.Printf("Querying metric. Start: %v, End: %v, query: %v", start, end, metric.Query)
	matchedSeries, err := client.QueryMetrics(start.Unix(), end.Unix(), metric.Query)
	if err != nil {
		log.Fatalf("fatal: %s\n", err)
	}
	applyOperation(metric, reduce(metric, matchedSeries[0]))
}

func applyOperation(metric Metric, value float64) {

	if value > metric.ScaleUp.Threshold {

		log.Printf("Value: %f > threshold: %f\tWould scale UP by: %d instances",
			value,
			metric.ScaleUp.Threshold,
			metric.ScaleUp.Count)
		group := getASG(metric.ScaleUp.GroupName, false)
		currentCapacity, _ := group.currentCapacity()
		log.Println("Current capacity: ", currentCapacity)
		group.scale(metric.ScaleUp.Count, false)
	} else if value < metric.ScaleDown.Threshold {

		log.Printf("Value: %f < threshold: %f\tWould scale DOWN by: %d instances",
			value,
			metric.ScaleDown.Threshold,
			metric.ScaleDown.Count)
		group := getASG(metric.ScaleUp.GroupName, false)
		currentCapacity, _ := group.currentCapacity()
		log.Println("Current capacity: ", currentCapacity)
		group.scale(metric.ScaleDown.Count, false)
	} else {
		log.Printf("Value does not match threshold: %f < %f < %f",
			metric.ScaleDown.Threshold, value, metric.ScaleUp.Threshold)
	}
}

func reduce(metric Metric, series datadog.Series) (value float64) {
	switch metric.Transform {
	case avgTransform:
		log.Println("applying avg transform")
		gen := dataPointValueGenerator(series.Points)
		for val := range gen {
			value = value + val
		}
		value = value / float64(len(series.Points))
	case minTransform:
		log.Println("applying min transform")
		gen := dataPointValueGenerator(series.Points)
		value := <-gen
		for val := range gen {
			value = math.Min(value, val)
		}
	case maxTransform:
		log.Println("applying max transform")
		gen := dataPointValueGenerator(series.Points)
		value := <-gen
		for val := range gen {
			value = math.Max(value, val)
		}
	case sumTransform:
		log.Println("applying sum transform")
		gen := dataPointValueGenerator(series.Points)
		for val := range gen {
			value = value + val
		}
	case lastTransform:
		log.Println("last transform")
		value = series.Points[len(series.Points)-1][1]
	case countTransform:
		log.Println("count transform")
		value = float64(len(series.Points))
	}
	return
}

func dataPointValueGenerator(points []datadog.DataPoint) chan (float64) {
	c := make(chan float64)

	go func() {
		for i := 0; i < len(points); i++ {
			c <- points[i][1] // get the "value" ignore the timestamp.
		}
		close(c)
	}()

	return c
}
