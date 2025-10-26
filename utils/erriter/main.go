package erriter

import (
	"iter"

	"github.com/jackc/pgx/v5"
)

type Iter[T any] struct {
	consumer int
	iter ErrIterFunc[T]
	err  error

	// For pgx
	curRow []any
	Close func()
	Trans  func(v T) []any
}

func (e *Iter[T]) takeConsumer(cType int) {
	if e.consumer != 0 && e.consumer != cType {
		panic("Consumer already taken")
	}
	e.consumer = cType
}

func (e *Iter[T]) Iter() iter.Seq[T] {
	e.takeConsumer(1)

	return func(yield func(T) bool) {
		e.err = e.iter(yield)
	}
}

func (e *Iter[T]) SafeClose() {
	if e.Close != nil {
		e.Close()
	}
}

func (e *Iter[T]) Err() error {
	return e.err
}

func (e *Iter[T]) Next() bool {
	e.takeConsumer(2)

	var row []any = nil

	e.err = e.iter(func (v T) bool {
		if e.Trans != nil {
			row = e.Trans(v)
		} else {
			row = []any{v}
		}

		return true
	})

	e.curRow = row
	if e.err != nil || row == nil {
		return false
	}

	return true
}

func (e *Iter[T]) Values() ([]any, error) {
	return e.curRow, e.err
}

var _ pgx.CopyFromSource = &Iter[string]{}

type ErrIterFunc[T any] func(yield func(T) bool) error

func New[T any](i ErrIterFunc[T]) *Iter[T] {
	return &Iter[T]{iter: i}
}

func Transform[T any](i iter.Seq[T]) *Iter[T] {
	return &Iter[T]{
		iter: func(yield func(T) bool) error {
			i(yield)
			return nil
		},
	}
}
