package circuit

import "testing"

func TestCounts(t *testing.T) {
	b := NewBucket(60)
	b.Fail()
	b.Success()
	b.Fail()
	b.Fail()
	b.Success()

	if f := b.Failures(); f != 3 {
		t.Fatalf("expected bucket to have 3 failures, got %d", f)
	}

	if s := b.Successes(); s != 2 {
		t.Fatalf("expected bucket to have 2 successes, got %d", s)
	}

	if r := b.ErrorRate(); r != 0.6 {
		t.Fatalf("expected bucket to have 0.6 error rate, got %f", r)
	}

	b.Reset()
	if f := b.Failures(); f != 0 {
		t.Fatalf("expected reset bucket to have 0 failures, got %d", f)
	}
	if s := b.Successes(); s != 0 {
		t.Fatalf("expected bucket to have 0 successes, got %d", s)
	}
}
