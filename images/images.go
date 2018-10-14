package images

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ttacon/chalk"

	"github.com/disintegration/imaging"
)

type jobReq struct {
	src    string
	picNum int
}

var grabLinkCounter = 0
var downloader = 0

func Get(links []string, folderName string, grayscale bool, start, limit, height, width int) {
	dSize := 100
	ich := make(chan string, len(links))
	lch := make(chan string, limit)
	jch := make(chan jobReq, limit)
	rch := make(chan string, dSize)
	ech := make(chan string)
	timeout := time.Duration(10 * time.Second)

	// create folder if does not exist
	mode := int64(0777)
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.MkdirAll(folderName, os.FileMode(mode))
	}
	// init workers
	for i := 1; i <= len(links); i++ {
		go loadList(i, ich, lch, limit, timeout)
	}
	for i := 1; i <= dSize; i++ {
		downloader++
		go download(i, jch, rch, ech, folderName, grayscale, height, width, timeout)
	}
	// start jobs
	for _, imgLnk := range links {
		fmt.Printf("load link from : %s\n", imgLnk)
		ich <- imgLnk
	}
	close(ich)

	for picNum := 1; picNum <= limit; picNum++ {
		url := <-lch
		fmt.Printf("scanning url: %s\tpicnum=%d\n", url, picNum)
		jch <- jobReq{
			src:    url,
			picNum: picNum,
		}
	}
	close(jch)

	var files []string
	for picNum := 1; picNum <= limit; picNum++ {
		select {
		case f := <-rch:
			fmt.Printf("%s Append file: %s %s\tpicNum=%d fileslen=%d chlen=%d\n", chalk.Cyan, chalk.Reset, f, picNum, len(files), len(rch))
			files = append(files, f)
		case <-ech:
		}

	}

	fmt.Println(chalk.Green, "Nb files :", chalk.Reset, len(files))
}

func loadList(id int, ich <-chan string, lch chan<- string, limit int, timeout time.Duration) {
	client := http.Client{
		Timeout: timeout,
	}
	defer client.CloseIdleConnections()
	for list := range ich {
		fmt.Printf("list loader %d: url: %s\n", id, list)
		// create a request to image-net
		req, err := http.NewRequest(http.MethodGet, list, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// scan through each line of the response body
		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			url := scanner.Text()
			fmt.Printf("list loader %d: image found: %s\t counter=%d\n", id, url, grabLinkCounter)
			lch <- url
			grabLinkCounter++
			if grabLinkCounter > limit {
				return
			}
		}
	}
}

func download(id int, jch <-chan jobReq, fch chan<- string, ech chan<- string, folderName string, grayscale bool, height, width int, timeout time.Duration) {
	client := http.Client{
		Timeout: timeout,
	}
	for j := range jch {
		fmt.Printf("worker %d: scanning line: %s\tpicnum=%d\n", id, j.src, j.picNum)

		// get the image associated with the link
		resp, e := client.Get(j.src)
		if e != nil || resp.StatusCode != http.StatusOK {
			ech <- j.src
			continue
		}
		defer resp.Body.Close()
		//open a file for writing
		filePath := filepath.Join(folderName, fmt.Sprintf("%d.jpg", j.picNum))
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("worker %d: error creating: %s\t picnum=%d\n", id, err, j.picNum)
			ech <- j.src
			continue
		}
		// Use io.Copy to just dump the response body to the file. This supports huge files
		n, err := io.Copy(file, resp.Body)
		if err != nil || n < 3000 {
			_ = os.Remove(filePath)
			ech <- j.src
			continue
		}
		file.Close()
		// open the file for image manipulation
		srcImg, err := imaging.Open(filePath)
		if srcImg == nil || err != nil {
			fmt.Printf("worker %d: error opening: %s \tpicnum=%d\n", id, err, j.picNum)
			ech <- j.src
			continue
		}
		// resize image to 100x100
		if height > 0 || width > 0 {
			srcImg = imaging.Resize(srcImg, width, height, imaging.Lanczos)
		}
		// convert image to grayscale
		if grayscale {
			srcImg = imaging.Grayscale(srcImg)
		}
		err = imaging.Save(srcImg, filePath)
		if err != nil {
			fmt.Printf("worker %d: error saving: %s \tpicnum=%d\n", id, err, j.picNum)
			ech <- j.src
			continue
		}
		fch <- filePath
	}
	downloader--
	fmt.Printf("%s# => Finish worker %d\tgoroutine=%d\n%s", chalk.Red, id, downloader, chalk.Reset)
}
