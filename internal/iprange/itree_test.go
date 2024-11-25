package iprange_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/danroc/geoblock/internal/iprange"
)

func TestInsert(t *testing.T) {
	tree := iprange.NewITree()

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.0.0"),
		End:   net.ParseIP("127.0.0.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.1.0"),
		End:   net.ParseIP("127.0.1.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.2.0"),
		End:   net.ParseIP("127.0.2.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.3.0"),
		End:   net.ParseIP("127.0.3.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.4.0"),
		End:   net.ParseIP("127.0.4.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.5.0"),
		End:   net.ParseIP("127.0.5.255"),
	})

	tree.Insert(iprange.Interval{
		Start: net.ParseIP("127.0.6.0"),
		End:   net.ParseIP("127.0.6.255"),
	})

	tree.Print()
	fmt.Printf("Tree height: %d\n", tree.Height())
}
