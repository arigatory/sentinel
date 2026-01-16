package main

import (
	"fmt"
	"runtime"
)

func main() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	fmt.Println("Alloc:", rtm.Alloc)
	fmt.Println("TotalAlloc:", rtm.TotalAlloc)
	fmt.Println("Sys:", rtm.Sys)
}
