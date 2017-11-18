package circuit

import (
	"testing"
)

func TestBreakerTripping(t *testing.T) {
	breaker := NewBreaker()

	if breaker.Tripped() {
		t.Fatal("expected breaker to not be tripped")
	}

	breaker.Trip()
	if !breaker.Tripped() {
		t.Fatal("expected breaker to be tripped")
	}

	breaker.Reset()
	if breaker.Tripped() {
		t.Fatal("expected breaker to not be tripped")
	}
}

func TestBreakerCounts(t *testing.T) {
	breaker := NewBreaker()

	breaker.Fail()
	if failures := breaker.Failures(); failures != 1 {
		t.Fatalf("expected failure count to be 1, got %d", failures)
	}

	breaker.Fail()
	if consecFailures := breaker.ConsecFailures(); consecFailures != 2 {
		t.Fatalf("expected 2 consecutive failures, got %d", consecFailures)
	}

	breaker.Success()
	if successes := breaker.Successes(); successes != 1 {
		t.Fatalf("expected success count to be 1, got %d", successes)
	}
	if consecFailures := breaker.ConsecFailures(); consecFailures != 0 {
		t.Fatalf("expected 0 consecutive failures, got %d", consecFailures)
	}

	breaker.Reset()
	if failures := breaker.Failures(); failures != 0 {
		t.Fatalf("expected failure count to be 0, got %d", failures)
	}
	if successes := breaker.Successes(); successes != 0 {
		t.Fatalf("expected success count to be 0, got %d", successes)
	}
	if consecFailures := breaker.ConsecFailures(); consecFailures != 0 {
		t.Fatalf("expected 0 consecutive failures, got %d", consecFailures)
	}
}

func TestErrorRate(t *testing.T) {
	cb := NewBreaker()
	if er := cb.ErrorRate(); er != 0.0 {
		t.Fatalf("expected breaker with no samples to have 0 error rate, got %f", er)
	}
}

func TestConsecFailureBreaker(t *testing.T) {
	option := &Options{
		ErrRate:         0.1,
		Sample:          100,
		ConsecFailTimes: 2,
		Interval:        5,
		BucketTimeout:   60,
	}
	breaker := NewBreakerWithOptions(option)

	if breaker.Tripped() {
		t.Fatal("expected breaker to be closed")
	}

	breaker.Fail()
	if breaker.Tripped() {
		t.Fatal("expected breaker to still be closed")
	}

	breaker.Fail()
	if !breaker.Tripped() {
		t.Fatal("expected breaker to be tripped")
	}

	breaker.Reset()
	if failures := breaker.Failures(); failures != 0 {
		t.Fatalf("expected reset to set failures to 0, got %d", failures)
	}
	if breaker.Tripped() {
		t.Fatal("expected threshold breaker to be closed")
	}
}

func TestErrorRateBreaker(t *testing.T) {
	option := &Options{
		ErrRate:         0.2,
		Sample:          10,
		ConsecFailTimes: 10,
		Interval:        5,
		BucketTimeout:   60,
	}
	breaker := NewBreakerWithOptions(option)

	if breaker.Tripped() {
		t.Fatal("expected breaker to be closed")
	}


	breaker.Fail()
	if breaker.Tripped() {
		t.Fatal("expected breaker to still be closed")
	}

	breaker.Fail()
	if breaker.Tripped() {
		t.Fatal("expected breaker to still be closed")
	}

	for i := 0; i < 8; i ++ {
		breaker.Success()
	}

	breaker.Fail()
	if !breaker.Tripped() {
		t.Fatal("expected breaker to be tripped")
	}

	breaker.Reset()
	if failures := breaker.Failures(); failures != 0 {
		t.Fatalf("expected reset to set failures to 0, got %d", failures)
	}
	if breaker.Tripped() {
		t.Fatal("expected threshold breaker to be closed")
	}
}


