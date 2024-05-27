package domain

import "fmt"

type Source struct {
	Filename string
	StartL   int
	EndL     int
}

func (s Source) Key() string {
	return fmt.Sprintf("%s-%d-%d", s.Filename, s.StartL, s.EndL)
}

type Clone struct {
	Filename string
	StartL   int
	EndL     int
	Distance float64
	// Sources indicate from which queries this co-change candidate was detected
	Sources []*Source
}

func (c Clone) Key() string {
	return fmt.Sprintf("%v-%d-%d", c.Filename, c.StartL, c.EndL)
}
