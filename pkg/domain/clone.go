package domain

import "fmt"

type Clone struct {
	Filename string
	StartL   int
	EndL     int
	Distance float64
}

func (c Clone) Key() string {
	return fmt.Sprintf("%v-%d-%d", c.Filename, c.StartL, c.EndL)
}
