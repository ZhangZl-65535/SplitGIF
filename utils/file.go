package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

func ReadFile(filepath string) []byte {
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println("ReadFile Open err:", err)
		return nil
	}

	defer f.Close()

	body, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("ReadFile ReadAll err:", err)
		return nil
	}
	return body
}

func WriteFile(filepath string, data []byte) {
	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		fmt.Println("WriteFile err:", err)
		return
	}
}
