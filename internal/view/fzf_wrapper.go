package view

import "github.com/ktr0731/go-fuzzyfinder"

type selectedFunc[T any] func(T) string
type previewFunc[T any] func(T) string

type FzfWrapper[T any] struct {
	collection []T
	selectedFunc[T]
	previewFunc[T]
	header string
}

func NewFinder[T any](
	collection []T,
	selectedFunc selectedFunc[T],
	previewFunc previewFunc[T],
	header string) FzfWrapper[T] {
	return FzfWrapper[T]{
		collection:   collection,
		selectedFunc: selectedFunc,
		previewFunc:  previewFunc,
		header:       header,
	}
}

func (w FzfWrapper[T]) Find() (T, error) {
	i, err := fuzzyfinder.Find(
		w.collection,
		func(i int) string {
			item := w.collection[i]
			return w.selectedFunc(item)
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i == -1 {
				return ""
			}

			return w.previewFunc(w.collection[i])
		}),
		fuzzyfinder.WithHeader(w.header),
	)

	if err != nil {
		return *new(T), err
	}

	return w.collection[i], nil
}
