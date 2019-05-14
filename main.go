package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/guiyomh/haar-training/images"
	"github.com/guiyomh/haar-training/samples"
	"github.com/guiyomh/haar-training/training"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {

	// CREATE COMMAND OPTION THAT TAKES FILEPATH TO POSITIVE FILE

	// 1. DOWNLOAD NEGATIVE BACKGROUND FILES
	nf := "negative_background.txt"
	links, err := readNegatives(nf)
	if err != nil {
		fmt.Printf("Error: Could not read the file: %s: %s", nf, err)
		os.Exit(1)
	}
	files := images.Get(links, "negatives", true, 1, 2000, 200, 200)

	// 2. GENERATE BG.TXT FILE FROM DOWNLOADED NEGATIVE (BACKGROUND) FILES

	//files, err := ioutil.ReadDir("negatives")
	// if err != nil {
	// 	fmt.Println("err reading dir:", err)
	// }
	var data string
	for _, file := range files {
		data += "negatives/" + file.Name() + "\n"
	}
	err = ioutil.WriteFile("bg.txt", []byte(data), 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// 3. CREATE POSITIVE SAMPLE VECTOR FILE
	// make training height and width variables
	// -w -h should be the width and height of your positive images
	createSampleCmdOptions := "-maxxangle 0.5 -maxyangle 0.5 maxzangle 0.5"
	samples.CreateSamples("manomano.jpg", "bg.txt", 1950, createSampleCmdOptions)

	// 5. TRAIN HAAR CASCADE FILE
	training.HaarCascade("data", 1800, 900, 20)
}

func readNegatives(file string) ([]string, error) {
	var links = make([]string, 10)
	f, err := os.Open(file)
	if err != nil {
		return links, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if len(url) > 0 && strings.HasPrefix(url, "http") {
			links = append(links, url)
		}
	}
	if scanner.Err() != nil {
		return links, scanner.Err()
	}
	return links, nil
}
