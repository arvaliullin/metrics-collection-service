package pool

import "sync"

// Resetter - тип с методом Reset() для сброса состояния перед возвратом в пул.
type Resetter interface {
	Reset()
}

// Pool - generic-контейнер для объектов типа T с методом Reset(); при возврате в пул объект сбрасывается.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New создаёт и возвращает указатель на Pool, для выделения новых экземпляров при пустом пуле используется newFunc.
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return newFunc() },
		},
	}
}

// Get возвращает объект из пула или новый из фабрики, если пул пуст.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put сбрасывает объект и возвращает его в пул для повторного использования.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
