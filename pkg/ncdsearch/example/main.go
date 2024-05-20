package main

import (
	"fmt"

	"github.com/salab/iccheck/pkg/ncdsearch"
)

func main() {
	clones := ncdsearch.Search([]byte("lo.SliceToMap(builds, func("), "/home/moto/tmp/NeoShowcase")
	fmt.Println(clones)
}
