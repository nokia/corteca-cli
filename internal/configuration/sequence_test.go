// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package configuration_test

import (
	"context"
	"github.com/nokia/corteca-cli/internal/configuration"
	"errors"
	"testing"
	"time"
)

// boolPtr returns a pointer to b, a convenience helper for SequenceCmd.IgnoreFailure.
func boolPtr(b bool) *bool { return &b }

// mockExecutor is a configurable test double for configuration.CommandExecutor.
type mockExecutor struct {
	beginErr error
	endErr   error

	beginCalled int
	endCalled   int

	// executeFunc is invoked on every ExecuteCommand call.
	// callIdx is 0-based. When nil, ExecuteCommand always returns (nil, nil).
	executeFunc func(callIdx int, ctx context.Context, cmd *configuration.SequenceCmd) (any, error)

	callCount        int
	capturedContexts []context.Context
}

func (m *mockExecutor) BeginSequence() error {
	m.beginCalled++
	return m.beginErr
}

func (m *mockExecutor) ExecuteCommand(ctx context.Context, cmd *configuration.SequenceCmd) (any, error) {
	m.capturedContexts = append(m.capturedContexts, ctx)
	idx := m.callCount
	m.callCount++
	if m.executeFunc != nil {
		return m.executeFunc(idx, ctx, cmd)
	}
	return nil, nil
}

func (m *mockExecutor) EndSequence() error {
	m.endCalled++
	return m.endErr
}

// alwaysFails returns an executeFunc that always returns the given error.
func alwaysFails(err error) func(int, context.Context, *configuration.SequenceCmd) (any, error) {
	return func(_ int, _ context.Context, _ *configuration.SequenceCmd) (any, error) {
		return nil, err
	}
}

// succeedsAfter returns an executeFunc that fails for the first n calls, then succeeds.
func succeedsAfter(n int, err error) func(int, context.Context, *configuration.SequenceCmd) (any, error) {
	return func(callIdx int, _ context.Context, _ *configuration.SequenceCmd) (any, error) {
		if callIdx < n {
			return nil, err
		}
		return "ok", nil
	}
}

// simpleStep is a convenience constructor for a SequenceCmd with only the fields that matter
// for a given test set.
func simpleStep(cmd string, ignoreFailure bool) configuration.SequenceCmd {
	return configuration.SequenceCmd{
		Cmd:           configuration.T(cmd),
		IgnoreFailure: boolPtr(ignoreFailure),
	}
}

// ---- SequenceMap.Execute: lifecycle tests ----------------------------------------

// TestExecute_UnknownSequence verifies that calling Execute with a name that does not exist
// in the map returns an error.
func TestExecute_UnknownSequence(t *testing.T) {
	sm := configuration.SequenceMap{"existing": {}}
	err := sm.Execute(&mockExecutor{}, "missing")
	if err == nil {
		t.Fatal("expected an error for unknown sequence name, got nil")
	}
}

// TestExecute_BeginAndEndSequence verifies that BeginSequence and EndSequence are each called
// exactly once when skipinit=false and the sequence completes successfully.
func TestExecute_BeginAndEndSequence(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {simpleStep("cmd", false)},
	}
	exec := &mockExecutor{}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.beginCalled != 1 {
		t.Errorf("BeginSequence: expected 1 call, got %d", exec.beginCalled)
	}
	if exec.endCalled != 1 {
		t.Errorf("EndSequence: expected 1 call, got %d", exec.endCalled)
	}
}

// TestExecute_EachStepCallsExecuteCommand verifies that ExecuteCommand is called once per step
// in the sequence.
func TestExecute_EachStepCallsExecuteCommand(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {
			simpleStep("cmd1", false),
			simpleStep("cmd2", false),
			simpleStep("cmd3", false),
		},
	}
	exec := &mockExecutor{}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.callCount != 3 {
		t.Errorf("ExecuteCommand: expected 3 calls (one per step), got %d", exec.callCount)
	}
}

// TestExecute_BeginSequenceError_AbortsBefore verifies that when BeginSequence returns an
// error, no steps are executed and EndSequence is not called.
func TestExecute_BeginSequenceError_AbortsBefore(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {simpleStep("cmd", false)},
	}
	exec := &mockExecutor{beginErr: errors.New("begin failed")}

	if err := sm.Execute(exec, "seq"); err == nil {
		t.Fatal("expected error when BeginSequence fails, got nil")
	}
	if exec.callCount != 0 {
		t.Errorf("ExecuteCommand: expected 0 calls after BeginSequence failure, got %d", exec.callCount)
	}
	if exec.endCalled != 0 {
		t.Errorf("EndSequence: expected 0 calls after BeginSequence failure, got %d", exec.endCalled)
	}
}

// TestExecute_StepFailure_EndSequenceSkipped verifies that when a step fails (and
// IgnoreFailure=false), Execute returns an error and EndSequence is NOT called.
func TestExecute_StepFailure_EndSequenceSkipped(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {simpleStep("cmd", false)},
	}
	exec := &mockExecutor{executeFunc: alwaysFails(errors.New("step failed"))}

	if err := sm.Execute(exec, "seq"); err == nil {
		t.Fatal("expected error for failed step, got nil")
	}
	if exec.endCalled != 0 {
		t.Errorf("EndSequence: expected 0 calls after step failure, got %d", exec.endCalled)
	}
}

// TestExecute_EndSequenceError_Propagates verifies that an error returned by EndSequence
// is propagated to the caller.
func TestExecute_EndSequenceError_Propagates(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {simpleStep("cmd", false)},
	}
	exec := &mockExecutor{endErr: errors.New("end failed")}

	if err := sm.Execute(exec, "seq"); err == nil {
		t.Fatal("expected error when EndSequence fails, got nil")
	}
}

// ---- Retries tests ---------------------------------------------------------------

// TestExecute_Retries_ExhaustsAllAttempts verifies that when a step always fails,
// ExecuteCommand is called exactly Retries+1 times before the sequence is aborted.
func TestExecute_Retries_ExhaustsAllAttempts(t *testing.T) {
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Retries:       2,
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}
	exec := &mockExecutor{executeFunc: alwaysFails(errors.New("fail"))}

	if err := sm.Execute(exec, "seq"); err == nil {
		t.Fatal("expected error after retries exhausted, got nil")
	}
	if exec.callCount != 3 {
		t.Errorf("ExecuteCommand: expected 3 calls (1 initial + 2 retries), got %d", exec.callCount)
	}
}

// TestExecute_Retries_SucceedsOnRetry verifies that the sequence completes successfully when
// a transient failure is resolved on a subsequent retry.
func TestExecute_Retries_SucceedsOnRetry(t *testing.T) {
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Retries:       2,
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}
	exec := &mockExecutor{executeFunc: succeedsAfter(1, errors.New("transient"))}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if exec.callCount != 2 {
		t.Errorf("ExecuteCommand: expected 2 calls (fail then succeed), got %d", exec.callCount)
	}
}

// ---- IgnoreFailure tests ---------------------------------------------------------

// TestExecute_IgnoreFailure_SkipsRetries verifies that when IgnoreFailure=true the command
// is executed exactly once, regardless of how many Retries are configured.
func TestExecute_IgnoreFailure_SkipsRetries(t *testing.T) {
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Retries:       2,
		IgnoreFailure: boolPtr(true),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}
	exec := &mockExecutor{executeFunc: alwaysFails(errors.New("fail"))}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("expected no error with IgnoreFailure=true, got: %v", err)
	}
	if exec.callCount != 1 {
		t.Errorf(
			"ExecuteCommand: expected exactly 1 call when IgnoreFailure=true (retries must be skipped), got %d",
			exec.callCount,
		)
	}
}

// TestExecute_IgnoreFailure_SequenceContinues verifies that execution advances to the next
// step after a step with IgnoreFailure=true fails.
func TestExecute_IgnoreFailure_SequenceContinues(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {
			{Cmd: configuration.T("fail-step"), IgnoreFailure: boolPtr(true)},
			{Cmd: configuration.T("ok-step"), IgnoreFailure: boolPtr(false)},
		},
	}
	exec := &mockExecutor{
		executeFunc: func(callIdx int, _ context.Context, _ *configuration.SequenceCmd) (any, error) {
			if callIdx == 0 {
				return nil, errors.New("ignored failure")
			}
			return "ok", nil
		},
	}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("expected sequence to complete successfully, got: %v", err)
	}
	if exec.callCount != 2 {
		t.Errorf("ExecuteCommand: expected 2 calls (both steps), got %d", exec.callCount)
	}
}

// TestExecute_IgnoreFailure_False_StopsOnError verifies that when IgnoreFailure=false and a
// step fails, the following steps are not executed.
func TestExecute_IgnoreFailure_False_StopsOnError(t *testing.T) {
	sm := configuration.SequenceMap{
		"seq": {
			simpleStep("fail-step", false),
			simpleStep("should-not-run", false),
		},
	}
	exec := &mockExecutor{
		executeFunc: func(callIdx int, _ context.Context, _ *configuration.SequenceCmd) (any, error) {
			if callIdx == 0 {
				return nil, errors.New("step failed")
			}
			return "ok", nil
		},
	}

	if err := sm.Execute(exec, "seq"); err == nil {
		t.Fatal("expected error for failed step, got nil")
	}
	if exec.callCount != 1 {
		t.Errorf("ExecuteCommand: expected 1 call (sequence must stop on failure), got %d", exec.callCount)
	}
}

// ---- Delay tests -----------------------------------------------------------------

// TestExecute_Delay_AppliedAfterAttempt verifies that a non-zero Delay causes an observable
// pause after ExecuteCommand returns.
func TestExecute_Delay_AppliedAfterAttempt(t *testing.T) {
	const delay = 20 * time.Millisecond
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Delay:         delay,
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}
	exec := &mockExecutor{}

	start := time.Now()
	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed < delay {
		t.Errorf("expected at least %v elapsed due to Delay, got %v", delay, elapsed)
	}
}

// TestExecute_Delay_AppliedBetweenRetries verifies that the Delay is applied after each
// failed attempt when retrying, producing a cumulative pause.
func TestExecute_Delay_AppliedBetweenRetries(t *testing.T) {
	const (
		delay   = 10 * time.Millisecond
		retries = 2
	)
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Retries:       retries,
		Delay:         delay,
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}
	// Succeed on the last attempt so the sequence finishes without error.
	exec := &mockExecutor{executeFunc: succeedsAfter(retries, errors.New("transient"))}

	start := time.Now()
	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Each of the three attempts is followed by a Delay sleep, so the total must be
	// at least retries+1 × delay.
	minExpected := delay * (retries + 1)
	if elapsed := time.Since(start); elapsed < minExpected {
		t.Errorf("expected at least %v elapsed (%d attempts × %v delay), got %v",
			minExpected, retries+1, delay, elapsed)
	}
}

// ---- Timeout tests ---------------------------------------------------------------

// TestExecute_Timeout_ContextDeadlineSet verifies that when Timeout is configured, the
// context passed to ExecuteCommand carries a matching deadline.
func TestExecute_Timeout_ContextDeadlineSet(t *testing.T) {
	const timeout = 500 * time.Millisecond
	cmd := configuration.SequenceCmd{
		Cmd:           configuration.T("cmd"),
		Timeout:       timeout,
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}

	var (
		capturedAt time.Time
		deadline   time.Time
		deadlineOk bool
	)
	exec := &mockExecutor{
		executeFunc: func(_ int, ctx context.Context, _ *configuration.SequenceCmd) (any, error) {
			capturedAt = time.Now()
			deadline, deadlineOk = ctx.Deadline()
			return nil, nil
		},
	}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deadlineOk {
		t.Fatal("context has no deadline; expected one to be set from Timeout")
	}
	// At the moment ExecuteCommand was entered, the remaining time must be positive
	// and no greater than the configured timeout.
	remaining := deadline.Sub(capturedAt)
	if remaining <= 0 || remaining > timeout {
		t.Errorf("context deadline remaining at capture time = %v; want in (0, %v]", remaining, timeout)
	}
}

// TestExecute_DefaultTimeout_UsedWhenUnset verifies that when Timeout is zero, the context
// deadline falls back to DefaultMaxTimeout (5 minutes).
func TestExecute_DefaultTimeout_UsedWhenUnset(t *testing.T) {
	cmd := configuration.SequenceCmd{
		Cmd: configuration.T("cmd"),
		// Timeout intentionally left at zero — should fall back to DefaultMaxTimeout.
		IgnoreFailure: boolPtr(false),
	}
	sm := configuration.SequenceMap{"seq": {cmd}}

	exec := &mockExecutor{
		executeFunc: func(_ int, ctx context.Context, _ *configuration.SequenceCmd) (any, error) {
			dl, ok := ctx.Deadline()
			if !ok {
				t.Error("context has no deadline; expected DefaultMaxTimeout to be applied")
				return nil, nil
			}
			remaining := time.Until(dl)
			// Allow a generous 1-second tolerance: the deadline must be within
			// (DefaultMaxTimeout-1s, DefaultMaxTimeout] from now.
			lo := configuration.DefaultMaxTimeout - time.Second
			hi := configuration.DefaultMaxTimeout
			if remaining < lo || remaining > hi {
				t.Errorf("default timeout: remaining = %v, want in [%v, %v]", remaining, lo, hi)
			}
			return nil, nil
		},
	}

	if err := sm.Execute(exec, "seq"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
