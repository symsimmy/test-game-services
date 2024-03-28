package prinitt

import (
	"fmt"
	"github.com/sasha-s/go-deadlock"
)

var (
	mutext deadlock.Mutex
)

func Printsss() {
	mutext.Lock()
	defer mutext.Unlock()
	fmt.Sprintf("%+v", "11111111111111111111")
}
