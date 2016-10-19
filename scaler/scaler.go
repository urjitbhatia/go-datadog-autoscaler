package scaler

import (
	"log"
	"math"
	"sort"
	"time"

	htime "github.com/urjitbhatia/gohumantime"
	"github.com/zorkian/go-datadog-api"
)

const (
	lastTransform  = "last"
	avgTransform   = "avg"
	minTransform   = "min"
	maxTransform   = "max"
	sumTransform   = "sum"
	countTransform = "count"

	scaleTypeUp   = "UP"
	scaleTypeDown = "DOWN"
)

type Scale struct {
	Count     int64
	Threshold float64
	Cooldown  bool
}

type Scales []Scale

type Metric struct {
	Name      string
	Query     string
	Period    string
	Transform string
	AwsRegion string
	GroupName string
	ScaleUp   Scales
	ScaleDown Scales
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
	if len(matchedSeries) == 0 {
		log.Fatalf("fatal: no matched series for given query\n")
	}
	applyOperation(metric, Reduce(metric, matchedSeries[0]))
}

func applyOperation(metric Metric, value float64) {

	scaleDirection := scaleTypeUp
	scale, ok := projectIntoScale(metric.ScaleUp, value, scaleDirection)
	if !ok {
		scaleDirection = scaleTypeDown
		scale, ok = projectIntoScale(metric.ScaleDown, value, scaleDirection)
	}

	if ok {
		log.Printf("\nSCALER: Value: %f matches threshold: %f\tWould scale %s by: %d instances",
			value, scale.Threshold, scaleDirection, scale.Count)

		group := getASG(metric.GroupName, metric.AwsRegion, false)
		currentCapacity, _ := group.currentCapacity()

		log.Printf("\nSCALER: Current capacity: %d Scaling to: %d", currentCapacity, currentCapacity+scale.Count)
		//		group.scale(scale.Count, false)

	} else {
		log.Printf("\nSCALER: Value does not match any scale threshold interval: %f", value)
	}
}

func projectIntoScale(scales Scales, value float64, scaleDirection string) (*Scale, bool) {

	var targetScale Scale
	var found bool

	log.Println("Checking scale direction", scaleDirection)
	switch scaleDirection {
	case scaleTypeUp:
		sort.Sort(scales)
		for _, scale := range scales {
			if value > scale.Threshold {
				targetScale = scale
				found = true
			} else {
				break
			}
		}
	case scaleTypeDown:
		sort.Sort(sort.Reverse(scales))
		for _, scale := range scales {
			if value < scale.Threshold {
				targetScale = scale
				found = true
			} else {
				break
			}
		}
	}
	return &targetScale, found
}

func Reduce(metric Metric, series datadog.Series) (value float64) {
	switch metric.Transform {
	case avgTransform:
		log.Println("applying avg transform")
		gen := UnzipDataPoints(series.Points)
		for val := range gen {
			value = value + val
		}
		value = value / float64(len(series.Points))
	case minTransform:
		log.Println("applying min transform")
		gen := UnzipDataPoints(series.Points)
		value = <-gen
		for val := range gen {
			value = math.Min(value, val)
		}
	case maxTransform:
		log.Println("applying max transform")
		gen := UnzipDataPoints(series.Points)
		value = <-gen
		for val := range gen {
			value = math.Max(value, val)
		}
	case sumTransform:
		log.Println("applying sum transform")
		gen := UnzipDataPoints(series.Points)
		for val := range gen {
			value = value + val
		}
	case lastTransform:
		log.Println("applying last transform")
		value = series.Points[len(series.Points)-1][1]
	case countTransform:
		log.Println("applying count transform")
		value = float64(len(series.Points))
	}
	return
}

func UnzipDataPoints(points []datadog.DataPoint) chan (float64) {
	c := make(chan float64)

	go func() {
		for i := 0; i < len(points); i++ {
			c <- points[i][1] // get the "value" ignore the timestamp.
		}
		close(c)
	}()

	return c
}

func emitEvent(title, text, resource string, client *datadog.Client) {
	event := &datadog.Event{}
	client.PostEvent(event)
}

func (slice Scales) Len() int {
	return len(slice)
}

func (slice Scales) Less(i, j int) bool {
	return slice[i].Threshold < slice[j].Threshold
}

func (slice Scales) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
