package email

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"time"
)

var maxBigInt = big.NewInt(math.MaxInt64)

func generateMessageID() (string, error) {
	t := time.Now().UnixNano()
	pid := os.Getpid()
	rint, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return "", err
	}
	h, err := os.Hostname()
	if err != nil {
		h = "localhost.localdomain"
	}
	id := fmt.Sprintf("<%d.%d.%d@%s>", t, pid, rint, h)
	return id, nil
}

func generateContentID(partname string) (string, error) {
	t := time.Now().UnixNano()
	pid := os.Getpid()
	rint, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("<%d.%d.%d@%s>", t, pid, rint, partname)
	return id, nil
}
