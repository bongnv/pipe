package task

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_taskImpl(t *testing.T) {
	t.Run("context-timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		mockTask := &taskImpl{
			done: make(chan struct{}),
		}

		cancel()
		err := mockTask.Wait(ctx)
		assert.EqualValues(t, context.Canceled, err)
	})

	t.Run("task-complete", func(t *testing.T) {
		called := false
		mockTask := &taskImpl{
			done: make(chan struct{}),
			doer: DoFunc(func(_ context.Context) error {
				called = true
				return nil
			}),
		}

		mockTask.start(context.Background())
		err := mockTask.Wait(context.Background())
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("task-panic", func(t *testing.T) {
		mockTask := &taskImpl{
			done: make(chan struct{}),
			doer: DoFunc(func(_ context.Context) error {
				panic("runtime panic")
			}),
		}

		assert.Panics(t, func() {
			mockTask.start(context.Background())
		})
	})

	t.Run("task-err", func(t *testing.T) {
		called := false
		mockTask := &taskImpl{
			done: make(chan struct{}),
			doer: DoFunc(func(_ context.Context) error {
				called = true
				return errors.New("runtime error")
			}),
		}

		mockTask.start(context.Background())
		err := mockTask.Wait(context.Background())
		assert.Error(t, err)
		assert.True(t, called)
	})

	t.Run("depepencies-error", func(t *testing.T) {
		called := false
		mockTask := &taskImpl{
			done: make(chan struct{}),
			doer: DoFunc(func(_ context.Context) error {
				called = true
				return nil
			}),
			dependencies: []Task{
				mockErrDependency(),
			},
		}

		mockTask.start(context.Background())
		err := mockTask.Wait(context.Background())
		assert.Error(t, err)
		assert.False(t, called)
	})
}

func mockErrDependency() Task {
	mock := &taskImpl{
		done: make(chan struct{}),
		err:  errors.New("random error"),
	}

	close(mock.done)
	return mock
}

func Test_Run(t *testing.T) {
	called := false
	doer := DoFunc(func(_ context.Context) error {
		called = true
		return nil
	})

	task1 := Run(context.Background(), doer)
	err := Wait(context.Background(), task1)
	assert.NoError(t, err)
	assert.True(t, called)
}
