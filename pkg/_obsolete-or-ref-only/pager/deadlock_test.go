package pager

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestRandomIntN(t *testing.T) {
	nn := make(map[int]int)
	for i := 0; i < 64; i++ {
		j := rand.Intn(32)
		nn[j]++
	}
	fmt.Println("random numbers:", nn)
}

func TestRandomInt(t *testing.T) {
	for i := 0; i < 64; i++ {
		j := rand.Int()
		fmt.Println(j)
	}
}

// run this with the -race flag
func TestUnsafeSafeMapAccess(t *testing.T) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}
		fmt.Printf(">> Recovered with: %v\n", v)
	}()

	useLock := true

	addLoop := func() {
		go loopAndSleepRand(0, 64, func() { setRandomKV(mm, useLock) })
	}

	getLoop := func() {
		go loopAndSleepRand(0, 64, func() { getRandomKV(mm, useLock) })
	}

	delLoop := func() {
		go loopAndSleepRand(0, 64, func() { delRandomKV(mm, useLock) })
	}
	go addLoop()
	go getLoop()
	go delLoop()

}

func loopAndSleepRand(beg, end int, fn func()) {
	for i := beg; i < end; i++ {
		fn()
		time.Sleep(time.Duration(rand.Intn(16)) * time.Millisecond)
	}
}

func sleepRand() {
	time.Sleep(time.Duration(rand.Intn(1<<7)) * time.Millisecond)
}

var mm = make(map[int]int)
var mu sync.RWMutex

func setRandomKV(m map[int]int, useLock bool) {
	if useLock {
		// hold write bpLatch
		mu.Lock()
		defer mu.Unlock()
	}
	j := rand.Intn(32)
	_, found := m[j]
	if !found {
		m[j] = j
		return
	}
	m[j]++
}

func getRandomKV(m map[int]int, useLock bool) int {
	if useLock {
		// hold read bpLatch
		mu.RLock()
		defer mu.RUnlock()
	}
	j := rand.Intn(32)
	v, found := m[j]
	if !found {
		return -1
	}
	return v
}

func delRandomKV(m map[int]int, useLock bool) {
	if useLock {
		// hold write bpLatch
		mu.Lock()
		defer mu.Unlock()
	}
	j := rand.Intn(32)
	_, found := m[j]
	if !found {
		return
	}
	delete(m, j)
}
