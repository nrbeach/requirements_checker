package main

import "os/exec"
import "bytes"
import "log"
import "fmt"

func main() {
	fmt.Println("Hello!")
	cmd := exec.Command("pip", "list")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Res: %q\n", out.String())
}

