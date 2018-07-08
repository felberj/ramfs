package ramfs

import (
	"testing"
)

func TestReadWrite(t *testing.T) {
	want := "hello world"
	node := &Node{}
	fd := &File{
		node: node,
	}
	n, err := fd.Write([]byte(want))
	if err != nil {
		t.Fatalf("write(%q)=%v", want, err)
	}
	if n != len(want) {
		t.Fatalf("write(%q) = %d, want %d", want, n, len(want))
	}
	fd2 := &File{
		node: node,
	}
	b := make([]byte, 1)
	n, err = fd2.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(b) {
		t.Fatalf("read() = %d, want %d", n, len(b))
	}
	bb := make([]byte, len(want)-len(b))
	n, err = fd2.Read(bb)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(bb) {
		t.Fatalf("read() = %d, want %d", n, len(bb))
	}
	b = append(b, bb...)
	if got := string(b); got != want {
		t.Fatalf("read() = %q, want %q", got, want)
	}
}
