package main

import (
	"os"

	"github.com/petems/tfvar-to-tfevar/cmd"
)

var version = "dev"

func main() {
	c, sync := cmd.New(os.Stdout, version)
	_ = c.Execute()

	sync()
}
