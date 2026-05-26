package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"slices"
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
	case "hash-object":
		filePath := os.Args[3]

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: could not open '%s' for reading: %s", filePath, err)
			break
		}

		fileContent, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			fmt.Fprint(os.Stderr, "fatal: unnable to read file\n")
			break
		}

		header := fmt.Appendf(nil, "blob %d\x00", len(fileContent))

		object := slices.Concat(header, fileContent)

		hashBytes := sha1.Sum(object)

		hash := hex.EncodeToString(hashBytes[:])

		dir := ".git/objects/" + hash[:2]
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: unnable to create object directory: %s\n", err)
			break
		}

		objectPath := dir + "/" + hash[2:]

		objectFile, err := os.OpenFile(objectPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: unnable to create object file: %s\n", err)
			break
		}
		defer objectFile.Close()

		w := zlib.NewWriter(objectFile)
		defer w.Close()

		_, err = w.Write(object)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: unnable to write to object file: %s\n", err)
			break
		}

		fmt.Printf("%s\n", hash)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
