package future_test

import (
	"context"
	"fmt"

	"github.com/knzm/future"
)

func genPrimes(s future.Sender) error {
	table := make(map[int]int)
	// q = 3, 5, 7, 9, ...
	for q := 3; ; q += 2 {
		p := table[q]
		if p == 0 {
			// q is a prime
			table[q*q] = q
			err := s.Send(q, 0)
			if err != nil {
				return err
			}
		} else {
			// q is a multipe of p
			delete(table, q)
			x := q + 2*p
			for table[x] != 0 {
				x += 2 * p
			}
			// x is the next multiple of p
			table[x] = p
		}
	}
	return nil
}

func PrintPrimes(n int) {
	ctx := context.Background()
	f := future.New(ctx, func(s future.Sender) error {
		return genPrimes(s)
	})
	defer f.Cancel()

	i := 0
	for v := range f.GetChannel() {
		if i >= n {
			f.Cancel()
			break
		}
		fmt.Println(v)
		i++
	}

	err := f.GetLastError()
	fmt.Println(err)
}

func ExamplePrimes() {
	PrintPrimes(20)
	// Output:
	// 3
	// 5
	// 7
	// 11
	// 13
	// 17
	// 19
	// 23
	// 29
	// 31
	// 37
	// 41
	// 43
	// 47
	// 53
	// 59
	// 61
	// 67
	// 71
	// 73
	// CANCELED
}
