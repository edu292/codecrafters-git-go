package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":

		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		object := os.Args[3]
		path := ".git/objects/" + object[:2] + "/" + object[2:]
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: Not a valid object name %s\n", object)
			break
		}

		r, err := zlib.NewReader(f)
		if err != nil {
			fmt.Fprint(os.Stderr, "fatal: unnable to read file\n")
			break
		}
		defer r.Close()

		p, err := io.ReadAll(r)
		if err != nil {
			fmt.Fprint(os.Stderr, "fatal: unnable to decompress object file\n")
		}

		_, content, found := bytes.Cut(p, []byte("\x00"))
		if !found {
			fmt.Fprintf(os.Stderr, "error: header for %s too long, exceeds 32 bytes\n", object)
			break
		}

		fmt.Printf("%s", content)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
