package iprange_test

import (
	"fmt"
	"testing"

	"github.com/danroc/geoblock/internal/iprange"
)

func TestInsert(t *testing.T) {
	tree := iprange.NewITree()

	tree.Insert(iprange.Interval{
		Start: 1,
		End:   3,
	}, 1)

	tree.Insert(iprange.Interval{
		Start: 4,
		End:   8,
	}, 2)

	tree.Insert(iprange.Interval{
		Start: 6,
		End:   10,
	}, 3)

	tree.Insert(iprange.Interval{
		Start: 11,
		End:   13,
	}, 4)

	tree.Insert(iprange.Interval{
		Start: 1,
		End:   13,
	}, 5)

	tree.Print()
	fmt.Printf("Tree height: %d\n", tree.Height())
	for i := 1; i <= 13; i++ {
		fmt.Printf("Query(%2d): %v\n", i, tree.Query(i))
	}
}
