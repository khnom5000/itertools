package itertools

import (
	"math"
	"reflect"
	"strings"
)

type Iterator chan interface{}

// Iter returns an Iterator for the iterables parameter
func Iter[T any](iterables []T) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		for _, value := range iterables {
			ch <- value
		}
	}()
	return
}

// Next goes to the next item within an Iterator
func Next(ch Iterator) any {
	return <-ch
}

// Repeat returns an Iterator which contains value parameter, size parameter amount of times
func Repeat(value any, size int) Iterator {
	s := make([]any, size)
	for i := range s {
		s[i] = value
	}
	return Iter(s)
}

// Zip iterates over multiple data objects in sync
func Zip[T any](iterables ...[]T) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		if ok := ensureSameLength(iterables); !ok {
			ch <- "all parameters must be of the same length"
			return
		}
		var toSend []any
		for index := range iterables[0] {
			toSend = nil
			for _, iterable := range iterables {
				toSend = append(toSend, iterable[index])
			}
			ch <- toSend
		}
	}()
	return
}

// Chain allows for multiple arrays of the same type to be iterated over
func Chain[T any](iterables ...[]T) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		for _, iterable := range iterables {
			for index := range iterable {
				ch <- iterable[index]
			}
		}
	}()
	return
}

// Count counts up from a certain number in an increment
func Count[T float32 | float64 | int](start, step T) (ch Iterator) {
	// consider changing step to uint
	ch = make(Iterator)
	go func() {
		defer close(ch)
		for {
			ch <- start
			start = start + step
		}
	}()
	return
}

// Cycle goes over a string seemingly forever
func Cycle(iterable string) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		for {
			letters := strings.SplitAfter(iterable, "")
			for _, letter := range letters {
				ch <- letter
			}
		}
	}()
	return
}

func Accumulate(iterable []int, operator string, start int) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		if start != 0 {
			ch <- start
		}
		toSend := iterable[0]
		ch <- toSend + start
		for _, element := range iterable[1:] {
			switch operator {
			case "add", "":
				toSend = toSend + element
			case "multiply":
				toSend = toSend * element
			case "power":
				toSend = int(math.Pow(float64(toSend), float64(element)))
			default:
				ch <- "not valid operator"
				return
			}
			ch <- toSend + start
		}
	}()
	return
}

func Tee[T []int | string](iterable T, n int) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		switch reflect.TypeOf(iterable).Kind() {
		case reflect.String:
			value := reflect.ValueOf(iterable).String()
			for len(value) != 0 {
				if len(value) < n {
					ch <- value
					return
				}
				ch <- value[0:n]
				value = value[n:]
			}
		case reflect.Array, reflect.Slice:
			value := reflect.ValueOf(iterable)
			for value.Len() != 0 {
				if value.Len() < n {
					ch <- value
					return
				}
				toSend := value.Slice(0, n)
				value = value.Slice(n, value.Len())
				ch <- toSend
			}
		}
	}()
	return
}

func Pairwise(iterable string) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		innerCh := Tee(iterable, 2)
		for value := range innerCh {
			ch <- value
		}

	}()
	return
}

// ensureSameLength ensures that all nested arrays are the same length
func ensureSameLength[T any](nestedList [][]T) bool {
	ch := Iter(nestedList)
	first := Next(ch)
	firstLength := reflect.ValueOf(first).Len()
	for nested := range ch {
		if reflect.ValueOf(nested).Len() != firstLength {
			return false
		}
	}
	return true
}

// Compress filters elements from data returning only those that have a corresponding element in selector that is true
func Compress[T any](data []T, selector []bool) (ch Iterator) {
	ch = make(Iterator)
	go func() {
		defer close(ch)
		for i, d := range data {
			if len(selector) > i && selector[i] {
				ch <- d
			}
		}
	}()
	return
}
