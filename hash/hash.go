package hash

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"net/url"
	"strings"
)

func charCodeAt(str string, index int) int {
	if index >= len(str) {
		return 0
	}
	return int(rune(str[index]))
}

func s(a string, b bool) []int {
	c := len(a)
	v := []int{}
	for i := 0; i < c; i += 4 {
		tmp := charCodeAt(a, i) | (charCodeAt(a, i+1) << 8) | (charCodeAt(a, i+2) << 16) | (charCodeAt(a, i+3) << 24)
		v = append(v, tmp)
	}
	if b {
		v = append(v, c)
	}
	return v
}

func l(a []int, b bool) string {
	d := len(a)
	c := (d - 1) << 2
	if b {
		m := a[d-1]
		if m < c-3 || m > c {
			return ""
		}
		c = m
	}
	res := []string{}
	for _, s := range a {
		item := string(rune(s&0xff)) + string(rune((s>>8)&0xff)) +
			string(rune((s>>16)&0xff)) + string(rune((s>>24)&0xff))
		res = append(res, item)
	}
	if b {
		return strings.Join(res, "")[0:c]
	} else {
		return strings.Join(res, "")
	}

}

// x encode
func X_encode(msg, key string) string {
	if msg == "" {
		return ""
	}
	v := s(msg, true)
	k := s(key, false)
	n := len(v) - 1
	z := v[n]
	y := v[0]
	c := 0x86014019 | 0x183639A0
	m := 0
	e := 0
	p := 0
	q := 6 + 52/(n+1)
	d := 0
	for ; 0 < q; q-- {
		d = (d + c) & (0x8CE0D9BF | 0x731F2640)
		e = d >> 2 & 3
		for p = 0; p < n; p++ {
			y = v[p+1]
			m = z>>5 ^ y<<2
			m += (y>>3 ^ z<<4) ^ (d ^ y)
			m += k[(p&3)^e] ^ z
			v[p] = (v[p] + m) & (0xEFB8D130 | 0x10472ECF)
			z = v[p]
		}
		y = v[0]
		m = z>>5 ^ y<<2
		m += (y>>3 ^ z<<4) ^ (d ^ y)
		m += k[(n&3)^e] ^ z
		v[n] = (v[n] + m) & (0xBB390742 | 0x44C6F8BD)
		z = v[n]
	}
	return l(v, false)
}

// md5 hash
func PwdHmd5(password, token string) string {
	hm := hmac.New(md5.New, []byte(token))
	hm.Write([]byte(password))
	hmd5 := hex.EncodeToString(hm.Sum(nil))
	return "{MD5}" + hmd5
}

// sha1 sum
func Checksum(data url.Values, token string) string {
	username := data.Get("username")
	password := data.Get("password")
	acid := data.Get("ac_id")
	ip := data.Get("ip")
	info := data.Get("info")
	strLists := []string{
		"",
		username,
		password[5:],
		acid, ip, "200", "1",
		info,
	}
	sumStr := strings.Join(strLists, token)
	sh := sha1.New()
	sh.Write([]byte(sumStr))
	return hex.EncodeToString(sh.Sum(nil))
}

// 加密信息
func GenInfo(data url.Values, token string) string {
	x_encode_json := map[string]interface{}{
		"username": data.Get("username"),
		"password": data.Get("password"),
		"ip":       data.Get("ip"),
		"acid":     data.Get("ac_id"),
		"enc_ver":  "srun_bx1",
	}

	x_encode_raw, err := json.Marshal(x_encode_json)
	if err != nil {
		logs.Debug(err)
		return ""
	}
	xen := string(x_encode_raw)
	x_encode_res := X_encode(xen, token)

	dict_key := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	dict_val := "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA="
	dict := map[string]string{}
	for idx, v := range dict_key {
		dict[string(v)] = dict_val[idx : idx+1]
	}
	b64_arr := []byte{}
	for _, c := range x_encode_res {
		b64_arr = append(b64_arr, byte(c))
	}
	b64_res := base64.StdEncoding.EncodeToString(b64_arr)
	target := ""
	for _, s := range b64_res {
		target += dict[string(s)]
	}
	return "{SRBX1}" + target
}
