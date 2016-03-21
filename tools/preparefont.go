package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
)

func prepareFont(fontFile, outputFile string) {

	dat, err := ioutil.ReadFile(fontFile)

	buf := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		log.Fatalf("Can't equire a new gzip writer: %s", err)
	}

	w.Comment = "Default font"
	w.Name = "defaultfont"

	w.Write(dat)
	err = w.Close()
	if err != nil {
		log.Fatalf("Can't close gzip writer: %s", err)
	}

	result := buf.Bytes()

	code := fmt.Sprintf(`package uik

import (
	"bytes"
	"compress/gzip"
	"io"
)

// DON'T EDIT! AUTOGENERATED CODE!
func defaultCode() []byte{
	gz, err := gzip.NewReader(
		bytes.NewBuffer(%+#v),
	)

	if err != nil {
		panic("Decompression failed: " + err.Error())
	}

	var b bytes.Buffer
	io.Copy(&b, gz)
	gz.Close()

	return b.Bytes()
}
`, result)
	err = ioutil.WriteFile(outputFile, []byte(code), 0644)
	if err != nil {
		log.Fatalf("Unable to save the code: %s", err)
	}
	var ratio float64
	datLen := len(dat)
	resLen := len(result)
	ratio = float64(len(dat)) / float64(len(result))
	fmt.Printf("Original size: %d\n", datLen)
	fmt.Printf("Compressed size: %d\n", resLen)
	if resLen > datLen {
		fmt.Println("\n\nHi! Unfortunally, you've found a bug.")
		fmt.Println("Font size become bigger after compression.")
		fmt.Printf("Please, submit this issue on our github repo:\n\t%s\n", githubRepoURL)
		fmt.Println("Note, that the result was saved.")
		fmt.Print("Thank you for using this kit.\n\n\n")
		return
	}
	fmt.Printf("Compressed by %.3f times\n", ratio)
	fmt.Printf("That saves you %d bytes\n", datLen-resLen)
}