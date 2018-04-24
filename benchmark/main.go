package main

import (
	"encoding/hex"
	"fmt"
)

func EncodeA(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func EncodeB(b []byte) string {
	return hex.EncodeToString(b)
}
