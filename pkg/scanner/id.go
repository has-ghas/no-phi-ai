package scanner

import (
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

func MakeDocumentID(hash plumbing.Hash, offset int, text_bytes []byte) (id string) {
	string_to_hash := strings.Join(
		[]string{
			hash.String(),
			strconv.Itoa(offset),
			string(text_bytes),
		},
		DelimitDocumentID,
	)

	hasher := sha1.New()
	hasher.Write([]byte(string_to_hash))
	id = base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return
}
