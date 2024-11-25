package pipeline

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func Link(p *Pipeline) string {
	return fmt.Sprint(RenderServer, Encoded(p.Diagram()))
}

func Encoded(diagram string) string {
	raw := []byte(diagram)
	compressed := deflate(raw)

	return base64Encode(compressed)
}

func deflate(input []byte) []byte {
	var b bytes.Buffer

	w, _ := zlib.NewWriterLevel(&b, zlib.BestCompression)
	_, _ = w.Write(input)
	_ = w.Close()

	return b.Bytes()
}

func base64Encode(input []byte) string {
	var buffer bytes.Buffer

	inputLength := len(input)

	for i := 0; i < 3-inputLength%3; i++ {
		input = append(input, byte(0))
	}

	for i := 0; i < inputLength; i += 3 {
		b1, b2, b3, b4 := input[i], input[i+1], input[i+2], byte(0)

		b4 = b3 & hSixtyThree
		b3 = ((b2 & hFifteen) << 2) | (b3 >> 6)
		b2 = ((b1 & hThree) << 4) | (b2 >> 4)
		b1 >>= 2

		for _, b := range []byte{b1, b2, b3, b4} {
			buffer.WriteByte(mapper[b])
		}
	}

	return encrypt(buffer)
}

func encrypt(buffer bytes.Buffer) string {
	plaintext := buffer.Bytes()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return fmt.Sprintf("%x", ciphertext)
}
