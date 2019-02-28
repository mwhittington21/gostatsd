package statsd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atlassian/gostatsd"
)

func Test_searchWhichBucket(t *testing.T) {
	tests := []struct {
		name    string
		buckets []gostatsd.BucketBounds
		v       float64
		want    int
	}{
		{name: "foo1", buckets: makeBounds(1), v: 0, want: 1},
		{name: "foo2", buckets: makeBounds(10, 20, 30), v: 21, want: 30},
		{name: "foo3", buckets: makeBounds(10, 20, 30), v: 30, want: PosInfinityBucketLimit},
		{name: "foo4", buckets: makeBounds(), v: 21, want: PosInfinityBucketLimit},
		{name: "foo5", buckets: makeBounds(10, 20, 30), v: 44, want: PosInfinityBucketLimit},
		{name: "foo7", buckets: makeBounds(10, 20, 30), v: 5, want: 10},
		{name: "foo8", buckets: makeBounds(10, 20, 30), v: 11, want: 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := searchWhichBucket(tt.buckets, tt.v); got.Max != tt.want {
				t.Errorf("searchWhichBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Timer(values ...float64) gostatsd.Timer {
	return gostatsd.Timer{
		Values:  values,
		Tags:    []string{PercentileBucketsMarkerTag, BucketPrefix + PercentileBucketsPow4Algorithm},
		Buckets: map[gostatsd.BucketBounds]int{},
	}
}

func TestAggregatePercentiles(t *testing.T) {
	//PercentileBucketsPow4Algorithm:          {4, 16, 64, 256, 1024, 4096, 16384},

	tests := []struct {
		name  string
		timer gostatsd.Timer
		want  map[gostatsd.BucketBounds]int
	}{
		{
			name:  "foo",
			timer: Timer(1, 10, 11, 12, 500, 1000, 1023, 1024, 1025, 4000, 16384),
			want: map[gostatsd.BucketBounds]int{
				// We include the zero-sized buckets in the comments here to make the test
				// more understandable.  We do not send save zero buckets in the map though.
				gostatsd.BucketBounds{0, 4}:  1, // 1
				gostatsd.BucketBounds{4, 16}: 3, // 10, 11, 12
				//64:    0,
				//256:   0,
				gostatsd.BucketBounds{256, 1024}:  3, // 500, 1000, 1023
				gostatsd.BucketBounds{1024, 4096}: 3, // 1024, 1025, 4000
				//16384:  0,
				gostatsd.BucketBounds{16384, PosInfinityBucketLimit}: 1, // 16384
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buckets := AggregatePercentiles(tt.timer)
			assert.Equal(t, tt.want, buckets)
		})
	}
}
