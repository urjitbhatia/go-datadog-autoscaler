package scaler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	scaler "github.com/urjitbhatia/go-datadog-autoscaler/scaler"
	datadog "github.com/zorkian/go-datadog-api"
)

var _ = Describe("Scaler", func() {
	Describe("Test value reduce functions", func() {
		Context("reduce", func() {
			metric := scaler.Metric{"TestMetric", "TestQuery",
				"TestPeriod", "",
				"TestRegion", "TestGroupName",
				nil, nil,
			}
			series := datadog.Series{}
			series.Points = []datadog.DataPoint{{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}}

			It("averages data points", func() {
				metric.Transform = "avg"
				Expect(scaler.Reduce(metric, series)).To(Equal(float64(3)))
			})
			It("gets the max of data points", func() {
				metric.Transform = "max"
				series.Points = append(series.Points, datadog.DataPoint{0, 6})
				Expect(scaler.Reduce(metric, series)).To(Equal(float64(6)))
			})
			It("gets the min of data points", func() {
				metric.Transform = "min"
				Expect(scaler.Reduce(metric, series)).To(Equal(float64(1)))
			})
			It("gets the last of data points", func() {
				metric.Transform = "last"
				series.Points = append(series.Points, datadog.DataPoint{0, 7})
				Expect(scaler.Reduce(metric, series)).To(Equal(float64(7)))
			})
			It("gets the count of data points", func() {
				metric.Transform = "count"
				series.Points = append(series.Points, datadog.DataPoint{0, 8})
				Expect(scaler.Reduce(metric, series)).To(Equal(float64(8)))
			})
		})

		Context("generate data points", func() {
			series := datadog.Series{}
			series.Points = []datadog.DataPoint{{0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}}

			It("unzip datapoint metric", func() {
				gen := scaler.UnzipDataPoints(series.Points)
				var values []float64
				for v := range gen {
				    values = append(values, v)
				}
				Expect(len(values)).To(Equal(5))
				Expect(values[2]).To(Equal(float64(3)))
			})
		})
	})
})
