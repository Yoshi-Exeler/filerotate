package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	for {
		appendWithRotate("myfile.txt", "THIS IS A LINE\n", 10, 100)
		time.Sleep(time.Millisecond * 50)
	}

}

func appendWithRotate(fileName string, content string, maxrot int, maxsize int) {
	// open a file handle with the appropriate options
	fileHandle, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer fileHandle.Close()
	// append the content to the file
	_, err = fileHandle.WriteString(content)
	if err != nil {
		panic(err)
	}
	info, err := fileHandle.Stat()
	if err != nil {
		panic(err)
	}
	// check if we exceeded the target file size
	if info.Size() > int64(maxsize) {
		fmt.Printf("target size exceeded size=%v target=%v\n", info.Size(), maxsize)
		// perform rotation
		rotate(fileName, content, maxrot, maxsize)
	}
}

func rotate(path string, content string, maxrot int, maxsize int) {
	fmt.Println("begin rotation")
	// first read all entries from the dir of the base file
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		panic(fmt.Sprintf("could not read directory with error %v\n", err))
	}
	// preallocate our ordered list in the size of maxrot
	orderedList := make([]string, maxrot)
	// sort the entries into the list slots
	for _, entry := range entries {
		// if the entry is not a directory and its name contains the base files name its a candidate
		if !entry.IsDir() && strings.Contains(entry.Name(), filepath.Base(path)) {
			idx, err := getRotationIndex(entry.Name())
			if err != nil {
				fmt.Printf("skipping %v because %v\n", entry.Name(), err)
				continue
			}
			// if the current entry is out of bounds for the current list, delete it
			if idx > len(orderedList)-1 {
				err := os.Remove(filepath.Join(filepath.Dir(path), entry.Name()))
				if err != nil {
					panic(fmt.Sprintf("file %v could not be deleted with error %v", entry.Name(), err))
				}
				fmt.Printf("file %v was deleted because it was out of bounds\n", entry.Name())
				continue
			}
			// save the index to the ordered list
			orderedList[idx] = entry.Name()
			fmt.Printf("saving list[%v] = %v\n", idx, entry.Name())
		}
	}
	// if there is an entry in the last slot, delete it
	if len(orderedList) > 0 && len(orderedList[len(orderedList)-1]) > 0 {
		err := os.Remove(filepath.Join(filepath.Dir(path), orderedList[len(orderedList)-1]))
		if err != nil {
			panic(fmt.Sprintf("file %v could not be deleted with error %v", orderedList[len(orderedList)-1], err))
		}
		fmt.Printf("file %v was deleted to make space for rotation\n", orderedList[len(orderedList)-1])
	}
	// shrink the list by one
	orderedList = orderedList[:len(orderedList)-1]
	fmt.Printf("shrunken list %v\n", orderedList)
	// reverse the list
	for i, j := 0, len(orderedList)-1; i < j; i, j = i+1, j-1 {
		orderedList[i], orderedList[j] = orderedList[j], orderedList[i]
	}
	fmt.Printf("reversed list %v\n", orderedList)
	// traverse the list and rename the files
	for index, entry := range orderedList {
		if len(entry) == 0 {
			fmt.Sprintf("skip index=%v entry=%v\n", index, entry)
			continue
		}
		// get the index of the current entry (cant error here)
		idx, _ := getRotationIndex(entry)
		// rename the file
		oldName := filepath.Join(filepath.Dir(path), entry)
		newName := filepath.Join(filepath.Dir(path), filepath.Base(path)+fmt.Sprintf(".%v", idx+1))
		err := os.Rename(oldName, newName)
		if err != nil {
			panic(fmt.Sprintf("could not rename %v to %v with error %v", oldName, newName, err))
		}
		fmt.Printf("rename %v >> %v\n", oldName, newName)
	}
	// rename the base file to 0
	os.Rename(path, filepath.Join(filepath.Dir(path), filepath.Base(path)+".0"))
}

// getRotationIndex returns the last dot split numeric segment in the file name
// for examplle base.json.5 returns 5
func getRotationIndex(fileName string) (int, error) {
	// split by dot
	parts := strings.Split(fileName, ".")
	// reverse the segments
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	// find the last segment
	for _, part := range parts {
		number, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		return int(number), nil
	}
	return 0, fmt.Errorf("filename %v contains no numeric segment", fileName)
}
