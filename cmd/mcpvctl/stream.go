package main

import "io"

func watchStream[T any](recv func() (T, error), handle func(T) error) error {
	for {
		item, err := recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := handle(item); err != nil {
			return err
		}
	}
}
