// @Author Adrian.Wang 2024/7/29 下午7:09:00
package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 use md5 to encrypt strings
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
