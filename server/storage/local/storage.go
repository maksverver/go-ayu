package local

import "bytes"
import "fmt"
import "io/ioutil"
import "path"

type LocalStorage struct {
	BaseDir string
}

func escape(bs []byte, buf *bytes.Buffer) {
	for _, b := range bs {
		switch {
		case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9', b == '_':
			buf.WriteByte(b)
		default:
			buf.WriteString(fmt.Sprintf("%%%02x", int(b)))
		}
	}
}

func (ls *LocalStorage) filename(kind string, key []byte) string {
	var buf bytes.Buffer
	escape([]byte(kind), &buf)
	buf.WriteRune('-')
	escape(key, &buf)
	buf.WriteString(".data")
	return path.Join(ls.BaseDir, buf.String())
}

func (ls *LocalStorage) Save(kind string, key, value []byte) error {
	return ioutil.WriteFile(ls.filename(kind, key), value, 0644)
}

func (ls *LocalStorage) Load(kind string, key []byte) ([]byte, error) {
	return ioutil.ReadFile(ls.filename(kind, key))
}
