package domain

import (
	"reflect"
	"testing"
)

func TestIgnoreRule_matchFile(t *testing.T) {
	cases := []struct {
		name     string
		pattern  string
		filename string
		match    bool
	}{
		{
			"go 1",
			"\\.go$",
			"main.go",
			true,
		},
		{
			"go 2",
			"\\.go$",
			"pkg/domain/hello.go",
			true,
		},
		{
			"go 3",
			"\\.go$",
			"src/main/java/go/hello.java",
			false,
		},
		{
			"go 4",
			"\\.go",
			"main.go",
			true,
		},
		{
			"dist",
			"^dist/",
			"dist/compiled.js",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ig := IgnoreConfig{Files: []string{c.pattern}}
			r, err := ig.Compile()
			if err != nil {
				t.Fatal(err)
			}
			got := r.matchFile(c.filename)
			if got != c.match {
				t.Errorf("got %v, want %v", got, c.match)
			}
		})
	}
}

func TestIgnoreRule_matchContents(t *testing.T) {
	const testContent = `package main

import "fmt"
import yamlv3 "gopkg.in/yaml.v3"

import (
  "fmt"
  yamlv3 "gopkg.in/yaml.v3"
)

func main() {
  fmt.Println("Hello World")
}
`

	cases := []struct {
		name     string
		pattern  string
		content  string
		expected map[int]struct{}
	}{
		{
			"go 1",
			`^package .+$`,
			testContent,
			map[int]struct{}{
				0: {},
			},
		},
		{
			"go 2",
			`import (.+ )?".+"$`,
			testContent,
			map[int]struct{}{
				2: {},
				3: {},
			},
		},
		{
			"go 3",
			`^import \(\n(\s+(.+ )?".+"\n)*\)$`,
			testContent,
			map[int]struct{}{
				5: {},
				6: {},
				7: {},
				8: {},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			i := IgnoreConfig{Patterns: []string{c.pattern}}
			ii, err := i.Compile()
			if err != nil {
				t.Fatal(err)
			}
			got := ii.matchContents([]byte(c.content))
			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("got %v, want %v", got, c.expected)
			}
		})
	}
}

func TestIgnoreLineRule_CanSkip(t *testing.T) {
	ignoreLines := &IgnoreLineRule{
		IgnoreLines: map[int]struct{}{
			2:  {},
			3:  {},
			4:  {},
			15: {},
		},
		safeUntil: -1,
	}
	windowSize := 9

	canSkip, skipUntil := ignoreLines.CanSkip(0, windowSize)
	if canSkip != true {
		t.Errorf("got %v, want %v", canSkip, true)
	}
	if skipUntil != 12 {
		t.Errorf("got %v, want %v", skipUntil, 12)
	}

	canSkip, skipUntil = ignoreLines.CanSkip(13, windowSize)
	if canSkip != true {
		t.Errorf("got %v, want %v", canSkip, true)
	}
	if skipUntil != 23 {
		t.Errorf("got %v, want %v", skipUntil, 23)
	}

	canSkip, _ = ignoreLines.CanSkip(24, windowSize)
	if canSkip != false {
		t.Errorf("got %v, want %v", canSkip, false)
	}
	canSkip, _ = ignoreLines.CanSkip(25, windowSize)
	if canSkip != false {
		t.Errorf("got %v, want %v", canSkip, false)
	}
}

func TestIgnoreRules_Match(t *testing.T) {
	const testContent = `package main

import "fmt"

import (
  "fmt"
)

func main() {
  fmt.Println("Hello World")
}
`

	i := IgnoreConfigs{
		// Ignore whole file rule
		{
			Files: []string{"^dist/"},
		},
		// Ignore specific patterns rule
		{
			Files: []string{"\\.go$"},
			Patterns: []string{
				`^package .+$`,
			},
		},
		{
			Files: []string{"^main.go$"},
			Patterns: []string{
				`^import \(\n(\s+(.+ )?".+"\n)*\)$`,
			},
		},
	}
	c, err := i.Compile()
	if err != nil {
		t.Fatal(err)
	}

	skipEntireFile, ignoreRule := c.Match("dist/compiled.js", []byte("console.log('hello world!');"))
	if skipEntireFile != true {
		t.Errorf("got %v, want %v", skipEntireFile, true)
	}
	if ignoreRule != nil {
		t.Errorf("got %v, want %v", ignoreRule, nil)
	}

	// Second and third rule should match
	skipEntireFile, ignoreRule = c.Match("main.go", []byte(testContent))
	if skipEntireFile != false {
		t.Errorf("got %v, want %v", skipEntireFile, false)
	}
	if ignoreRule == nil {
		t.Errorf("got ignoreRule == nil, want non-nil")
	} else {
		wantIgnoreLines := map[int]struct{}{
			0: {},
			4: {},
			5: {},
			6: {},
		}
		if !reflect.DeepEqual(ignoreRule.IgnoreLines, wantIgnoreLines) {
			t.Errorf("got %v, want %v", ignoreRule.IgnoreLines, wantIgnoreLines)
		}
	}

	// Only the third rule should match
	skipEntireFile, ignoreRule = c.Match("pkg/foo.go", []byte(testContent))
	if skipEntireFile != false {
		t.Errorf("got %v, want %v", skipEntireFile, false)
	}
	if ignoreRule == nil {
		t.Errorf("got ignoreRule == nil, want non-nil")
	} else {
		wantIgnoreLines := map[int]struct{}{
			0: {},
		}
		if !reflect.DeepEqual(ignoreRule.IgnoreLines, wantIgnoreLines) {
			t.Errorf("got %v, want %v", ignoreRule.IgnoreLines, wantIgnoreLines)
		}
	}
}
