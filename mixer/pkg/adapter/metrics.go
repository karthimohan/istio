// Copyright 2017 The Istio Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"errors"
	"time"
)

// Metric kinds supported by mixer.
const (
	Gauge        MetricKind = iota // records instantaneous (non-cumulative) measurements
	Counter                        // records increasing cumulative values
	Distribution                   // aggregates values in buckets (values still reported un-aggregated)
)

type (
	// MetricsAspect handles metric reporting within the mixer.
	MetricsAspect interface {
		Aspect

		// Record directs a backend adapter to record the list of values
		// that have been generated from Report() calls.
		Record([]Value) error
	}

	// Value holds a single metric value that will be generated through
	// a Report() call to the mixer. It is synthesized by the mixer, based
	// on mixer config and the attributes passed to Report().
	Value struct {
		// The definition describing this metric.
		Definition *MetricDefinition
		// Labels provide metadata about the metric value. They are
		// generated from the set of attributes provided by Report().
		Labels map[string]interface{}
		// StartTime marks the beginning of the period for which the
		// metric value is being reported. For instantaneous metrics,
		// StartTime records the relevant instant.
		StartTime time.Time
		// EndTime marks the end of the period for which the metric
		// value is being reported. For instantaneous metrics, EndTime
		// will be set to the same value as StartTime.
		EndTime time.Time

		// The value of this metric; this should be accessed type-safely via value.String(), value.Bool(), etc.
		MetricValue interface{}
	}

	// MetricKind defines the set of known metrics types that can be generated
	// by the mixer.
	MetricKind int

	// MetricsBuilder builds instances of the Metrics aspect.
	MetricsBuilder interface {
		Builder

		// NewMetricsAspect returns a new instance of the Metrics aspect.
		NewMetricsAspect(env Env, config Config, metrics map[string]*MetricDefinition) (MetricsAspect, error)
	}

	// MetricDefinition provides the basic description of a metric schema
	// for which metrics adapters will be sent Values at runtime.
	MetricDefinition struct {
		// Name is the canonical name of the metric.
		Name string
		// Optional user-friendly name of the metric.
		DisplayName string
		// Optional user-friendly description of this metric.
		Description string
		// Kind provides type information about the metric.
		Kind MetricKind
		// Labels are the names of keys for dimensional data that will
		// be generated at runtime and passed along with metric values.
		Labels map[string]LabelType

		Buckets BucketDefinition
	}

	// BucketDefinition provides a common interface for the various types
	// of bucket definitions.
	BucketDefinition interface{}

	// LinearBuckets describes a linear sequence of buckets that all have
	// the same width (except for underflow and overflow).
	//
	// There are `Count + 2` (= N) buckets. The two additional
	// buckets are the underflow and overflow buckets.
	//
	// Bucket `i` has the following boundaries:
	//    Upper bound (0 <= i < N-1):     offset + (width * i).
	//    Lower bound (1 <= i < N):       offset + (width * (i - 1)).
	LinearBuckets struct {
		BucketDefinition

		// Count is the number of buckets in this bucket definition.
		// Must be greater than 0.
		Count int32

		// Width describes the size of each individual bucket. Must be
		// greater than 0.
		Width float64

		// Offset is the lower bound of the first specified bucket.
		Offset float64
	}

	// ExponentialBuckets describes an exponential sequence of buckets that
	// have a width that is proportional to the value of the lower bound.
	//
	// There are `Count + 2` (= N) buckets. The two additional
	// buckets are the underflow and overflow buckets.
	//
	// Bucket `i` has the following boundaries:
	//    Upper bound (0 <= i < N-1):     scale * (growth_factor ^ i).
	//    Lower bound (1 <= i < N):       scale * (growth_factor ^ (i - 1)).
	ExponentialBuckets struct {
		BucketDefinition

		// Count is the number of buckets in this bucket definition.
		// Must be greater than 0.
		Count int32

		// GrowthFactor controls the rate of increase in size of the
		// buckets. Must be greater than 1.
		GrowthFactor float64

		// Scale controls the relative size of the buckets. Must be
		// greater than 0.
		Scale float64
	}

	// ExplicitBuckets specifies a set of buckets with arbitrary widths.
	//
	// There are `size(bounds) + 1` (= N) buckets. Bucket `i` has the following
	// boundaries:
	//
	//    Upper bound (0 <= i < N-1):     bounds[i]
	//    Lower bound (1 <= i < N);       bounds[i - 1]
	//
	// If `Bounds` has only one element, then there are no finite buckets,
	// and that single element is the common boundary of the overflow and
	// underflow buckets.
	ExplicitBuckets struct {
		BucketDefinition

		// Bounds describes the explicit bounds of the buckets being
		// defined. Must be at least 1 element long and have
		// monotonically increasing values.
		Bounds []float64
	}
)

// String returns the string-valued metric value for a metrics.Value.
func (v Value) String() (string, error) {
	if v, ok := v.MetricValue.(string); ok {
		return v, nil
	}
	return "", errors.New("metric value is not a string")
}

// Bool returns the boolean metric value for a metrics.Value.
func (v Value) Bool() (bool, error) {
	if v, ok := v.MetricValue.(bool); ok {
		return v, nil
	}
	return false, errors.New("metric value is not a boolean")
}

// Int64 returns the int64-valued metric value for a metrics.Value.
func (v Value) Int64() (int64, error) {
	if v, ok := v.MetricValue.(int64); ok {
		return v, nil
	}
	return 0, errors.New("metric value is not an int64")
}

// Float64 returns the float64-valued metric value for a metrics.Value.
func (v Value) Float64() (float64, error) {
	if v, ok := v.MetricValue.(float64); ok {
		return v, nil
	}
	return 0, errors.New("metric value is not a float64")
}
