package main

import (
	"fmt"
	"log"
	"time"
)

// MakeBallast is a function that creates a large allocation of memory that can
// provide stability to the heap. Much like the nautical ballast its purpose is
// to provide desired draft and stability.
//
// The GC will trigger every time the heap size doubles. The heap size is the total
// size of allocations on the heap. Therefore, if a ballast of 10MB is allocated,
// the next GC will only trigger when the heap size grows to 20MB. At that point,
// there will be roughly 10MB of ballast + 10MB of other allocations.
//
// When the GC runs, the ballast will not be swept as garbage since we still hold a
// reference to it in our main function, and thus it is considered part of the live
// memory. Since most of the allocations in our application only exist for the short
// lifetime, most of the 10MB of allocation will get swept, reducing the heap back
// down to just over ~10MB again (i.e., the 10MB of ballast plus whatever in flight
// requests have allocations and are considered live memory.) Now, the next GC cycle
// will occur when the heap size (currently just larger than 10MB) doubles again.
//
// So in summary, the ballast increases the base size of the heap so that our GC
// triggers are delayed and the number of GC cycles over time is reduced.
//
func MakeBallast(b *[]byte, mb int) {
	// We use a byte array for the ballast, this is to ensure that we only add one
	// additional object to the mark phase. Since a byte array does not have any
	// pointers (other than the object itself), the GC can mark the entire object
	// in O(1) time.
	*b = make([]byte, mb<<20)
}

func main() {
	var b []byte
	MakeBallast(&b, 16)
	WaitForKey("", func() { log.Println("Later!") })
}

// WaitForKey waits for you to press any key to continue.
func WaitForKey(msg string, after func()) {
	if msg == "" {
		msg = "Press any key to continue..."
	}
	log.Printf(msg)
	if _, err := fmt.Scanln(); err != nil {
		log.Panicf("Scanln: %q\n", err)
	}
	if after != nil {
		defer after()
	}
}

func doAfter(n int) {
	fmt.Printf("I am running after the keypress (n=%d)\n", n)
}

func makeNoise() {
	noise := func() {
		for i := 0; i < 32; i++ {
			_ = make([]uint64, 255)
		}
	}
	for i := 0; i < 64; i++ {
		noise()
		time.Sleep(10 * time.Second)
	}
}
