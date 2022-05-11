package _pager

import (
	"fmt"
	"testing"
)

func TestPage_GetPageHeaderUnsafe(t *testing.T) {
	pg := newPage(1)
	fmt.Printf("page header:\n%s\n", pg.PageHeaderString())
}
