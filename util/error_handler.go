package util

import "log"

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf(msg, err)
	}
}
