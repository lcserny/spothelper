package spothelper

import (
	"bufio"
	"fmt"
	"os"
)

func ReadFile(file string) {
	/*	allBytes, err := ioutil.ReadFile(file)
		check(err)
		fmt.Print(string(allBytes) + "\n")*/

	openFile, err := os.Open(file)
	defer CloseFile(openFile)
	CheckError(err)

	/*	b1 := make([]byte, 5)
		r1, err := openFile.Read(b1)
		check(err)
		fmt.Printf("%d bytes: %s\n", r1, string(r1))

		s2, err := openFile.Seek(6, 0)
		check(err)
		b2 := make([]byte, 2)
		r2, err := openFile.Read(b2)
		check(err)
		fmt.Printf("%d bytes @ %d: %s\n", r2, s2, string(r2))*/

	/*	reader := bufio.NewReader(openFile)
		for line, _, _ := reader.ReadLine(); line != nil; {
			fmt.Printf("This is the line: %s", string(line))
		}*/

	scanner := bufio.NewScanner(openFile)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func CloseFile(file *os.File) {
	err := file.Close()
	CheckError(err)
}
