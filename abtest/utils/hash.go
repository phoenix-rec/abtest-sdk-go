package utils

import (
	"crypto/md5"
	"fmt"
	"strconv"
)

func HashIndex(expID, id string, totalCount uint64) (index uint64) {
	s := fmt.Sprintf("%x", md5.Sum([]byte(expID+id)))
	i, _ := strconv.ParseUint(s[24:], 16, 64)
	index = i % totalCount
	return
}
