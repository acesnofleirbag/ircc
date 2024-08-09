package guard

import (
	"log"
)

func Err(err error) {
    if err != nil {
        log.Fatalf("%+v", err)
    }
}

