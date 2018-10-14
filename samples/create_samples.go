package samples

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Opencv_createsamples -img watch5050.jpg -bg bg.txt -info info/info.lst -pngoutput info -maxxangle 0.5 -maxyangle -0.5 -maxzangle 0.5 -num 1950 -w 20 -h 20
func CreateSamples(posFile string, bgFile string, num int, createSampleCmdOptions string) {
	// if "info" doesn't exist create it
	mode := int64(0777)
	if _, err := os.Stat("info"); os.IsNotExist(err) {
		os.MkdirAll("info", os.FileMode(mode))
	}
	cmdName, err := exec.LookPath("opencv_createsamples")
	if err != nil {
		fmt.Println("can't find opencv_createsamples")
		return
	}
	cmdArgs := []string{
		"-img", posFile,
		"-bg", bgFile,
		//	"-info", "info/info.lst",
		"-w", "70",
		"-h", "70",
		createSampleCmdOptions,
		"-vec", "positives.vec",
		"-num", strconv.Itoa(num),
	}
	fmt.Println(cmdName, cmdArgs)
	_, err = exec.Command(cmdName, cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println("err:", err)
		return
	}
}
