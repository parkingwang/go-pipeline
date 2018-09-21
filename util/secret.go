package util

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

const (
	letters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randomSrc = rand.NewSource(time.Now().UnixNano())

func RandString(num int) string {
	bytes := make([]byte, num)
	for i, cache, remain := num-1, randomSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randomSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			bytes[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(bytes)
}

// 生成签名字符串
func GenerateAuthSign(secretKey string) (string, error) {
	txt := fmt.Sprintf("timestamp=%d&nonce=%s", time.Now().Unix(), RandString(6))
	sign := CreateSignText(txt, secretKey)
	return fmt.Sprintf("%s&sign=%s", txt, sign), nil
}

func CreateSignText(txt string, secretKey string) string {
	hmacx := hmac.New(sha1.New, []byte(secretKey))
	hmacx.Write([]byte(txt))
	sign := hmacx.Sum(nil)
	return fmt.Sprintf("%x", sign[:6])
}

func VerifySignature(sign string, tsTimeout time.Duration, secretKey string) (bool, error) {
	maps, err := url.ParseQuery(sign)
	if nil != err {
		return false, err
	}

	timestamp2 := maps["timestamp"]
	if 1 != len(timestamp2) {
		return false, errors.New("[timestamp] field is required")
	}

	// check timestamp
	unix, err := strconv.ParseInt(timestamp2[0], 10, 64)
	if nil != err {
		return false, err
	}

	du := time.Since(time.Unix(unix, 0))
	if math.Abs(du.Seconds()) > tsTimeout.Seconds() {
		return false, fmt.Errorf("timestamp[%d] is timeout: %s", unix, du.String())
	}

	nonce2 := maps["nonce"]
	if 1 != len(nonce2) {
		return false, errors.New("nonce field is required")
	}

	sign2 := maps["sign"]
	if 1 != len(sign2) {
		return false, errors.New("sign field is required")
	}

	txt := fmt.Sprintf("timestamp=%s&nonce=%s", timestamp2[0], nonce2[0])

	if CreateSignText(txt, secretKey) == sign2[0] {
		return true, nil
	} else {
		return false, errors.New("sign not match")
	}
}
