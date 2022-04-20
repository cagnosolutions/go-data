package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cagnosolutions/go-data/pkg/sorcerer"
)

func main() {
	// get new source file wizard
	s := sorcerer.NewSorcerer()
	RunParseExpression(s)
	RunParseFile(s)
	RunParseFile(s)
	RunParseFile(s)
	// RunParseFile2(s, "files/one.go")
	// RunParseFile2(s, "files/two.go")
	// RunParseFile2(s, "files/three.go")
	// RunParseDir(s)

	time.Sleep(5 * time.Second)
	fmt.Println(s.Structs)
}

func RunParseExpression(s *sorcerer.Sorcerer) {
	// parse the expression
	err := s.ParseExpression("package main; func add(a, b int) int { return a + b }")
	if err != nil {
		log.Fatalf("error parsing expression: %s\n", err)
	}
	// print parsed file or expression
	key := "expr"
	f, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	fmt.Printf("file: %s\n", f)
}

func RunParseFile(s *sorcerer.Sorcerer) {
	// parse the file
	filename := "_hello.go"
	err := s.ParseFile(filename)
	if err != nil {
		log.Fatalf("error parsing file %q: %s\n", filename, err)
	}
	// print parsed file or expression
	key := filename
	f, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	fmt.Printf("file: %s\n", f)
}

func RunParseFile2(s *sorcerer.Sorcerer, file string) {
	// parse the file
	filename := file
	err := s.ParseFile(filename)
	if err != nil {
		log.Fatalf("error parsing file %q: %s\n", filename, err)
	}
	// print parsed file or expression
	key := filename
	f, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	fmt.Printf("file: %s\n", f)
}

func RunParseDir(s *sorcerer.Sorcerer) {
	// parse the dir
	dir := "files"
	err := s.ParseDir(dir, nil)
	if err != nil {
		log.Fatalf("error parsing dir %q: %s\n", dir, err)
	}
	// print parsed files
	key := "one.go"
	f1, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	key = "two.go"
	f2, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	key = "three.go"
	f3, err := s.GetFile(key)
	if err != nil {
		log.Fatalf("error getting file %q: %s\n", key, err)
	}
	fmt.Printf("file: %s\n", f1)
	fmt.Printf("file: %s\n", f2)
	fmt.Printf("file: %s\n", f3)
}
