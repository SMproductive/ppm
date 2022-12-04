package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func main() {
	out := exec.Command("mkfifo", "foo")
	out.Run()
	a, _ := ioutil.ReadFile("foo")
	if string(a) == "hi\n" {
		fmt.Println("hi buddy!")
	} else {
		fmt.Println("didnt hear from you")
	}
	fmt.Println(string(a))

}
