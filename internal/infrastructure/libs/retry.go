package libs

import "time"

func Retry(fn func() error) {
	for i := 0; i < 3; i++ {
		err := fn()
		if err == nil {
			return
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
}
