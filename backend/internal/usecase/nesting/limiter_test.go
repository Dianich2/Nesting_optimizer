package nesting

import (
	"context"
	"crypto/sha256"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
	"time"
)

func TestComputeLimiterBoundsAndCancellation(t *testing.T) {
	limiter := NewComputeLimiter(1, 1)
	release, err := limiter.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		slot, err := limiter.Acquire(ctx)
		if err == nil {
			slot()
		}
		done <- err
	}()
	expires := time.Now().Add(time.Second)
	for len(limiter.admitted) != 2 {
		if time.Now().After(expires) {
			t.Fatal("request did not enter bounded queue")
		}
		time.Sleep(time.Millisecond)
	}
	if _, err := limiter.Acquire(t.Context()); err == nil {
		t.Fatal("unbounded admission")
	} else if failure, ok := apperror.As(err); !ok || failure.Code != apperror.AppCodeOverloaded {
		t.Fatalf("overload error: %v", err)
	}
	cancel()
	if err := <-done; err != context.Canceled {
		t.Fatalf("waiting cancellation: %v", err)
	}
	release()
	func() {
		defer func() { _ = recover() }()
		slot, err := limiter.Acquire(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer slot()
		panic("injected optimizer panic")
	}()
	ctx, cancel = context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	slot, err := limiter.Acquire(ctx)
	if err != nil {
		t.Fatalf("slot leaked after panic: %v", err)
	}
	slot()
}

func TestLegacyNestingFingerprint(t *testing.T) {
	input := RunNestingInput{Algorithm: optimizer.BaselineAlgorithm, UserID: 1, ProjectID: 1, ProjectSurfaceID: 1, Patterns: []RunNestingPatternInput{{ProjectPatternID: 1, Quantity: 1}}, AllowedRotations: []float64{0}}
	old := []byte(`{"Algorithm":"baseline","UserID":1,"ProjectID":1,"ProjectSurfaceID":1,"Patterns":[{"ProjectPatternID":1,"Quantity":1}],"AllowedRotations":[0],"KeepExisting":false}`)
	hash, err := Fingerprint(input)
	if err != nil || hash != sha256.Sum256(old) {
		t.Fatal("existing saved idempotency keys changed")
	}
}

func TestDuplicateResponseWaitersAreBoundedAndIndependent(t *testing.T) {
	limiter := NewComputeLimiter(1, 0)
	releases := make([]func(), 0, 32)
	for range 32 {
		release, err := limiter.AwaitResponse(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, release)
	}
	if _, err := limiter.AwaitResponse(t.Context()); err == nil {
		t.Fatal("unbounded duplicate waiters")
	} else if failure, ok := apperror.As(err); !ok || failure.Code != apperror.AppCodePending {
		t.Fatalf("duplicate overload is not distinguishable: %v", err)
	}
	compute, err := limiter.Acquire(t.Context())
	if err != nil {
		t.Fatalf("response waiters occupied optimizer capacity: %v", err)
	}
	compute()
	for _, release := range releases {
		release()
	}
	release, err := limiter.AwaitResponse(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	release()
}
