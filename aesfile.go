package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func EncryptFile(key string, filename string) error {
	plaintext, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	bKey := []byte(key)
	block, err := aes.NewCipher(bKey)
	if err != nil {
		return err
	}
	plaintext = PKCS5Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	bm := cipher.NewCBCEncrypter(block, iv)
	bm.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	f, err := os.Create(filename + ".enc")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(ciphertext))
	if err != nil {
		return err
	}

	return nil
}

func DecryptFile(key string, filename string) error {
	ciphertext, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	bKey := []byte(key)
	block, err := aes.NewCipher(bKey)
	if err != nil {
		return err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	bm := cipher.NewCBCDecrypter(block, iv)
	bm.CryptBlocks(ciphertext, ciphertext)
	ciphertext = PKCS5UnPadding(ciphertext)

	f, err := os.Create(filename + ".dec")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(ciphertext))
	if err != nil {
		return err
	}

	return nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n")
	}
	flag.Usage()
	os.Exit(1)
}

func errorAndExit(err error) {
	fmt.Fprintf(os.Stderr, err.Error())
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func main() {
	isEncrypt := flag.Bool("e", false, "encrypt")
	isDecrypt := flag.Bool("d", false, "decrypt")
	inputFile := flag.String("f", "", "filepath")
	key := flag.String("k", "", "key")

	flag.Parse()

	if *inputFile == "" {
		usageAndExit("Missing file")
	} else if *key == "" {
		usageAndExit("Missing key")
	} else if !*isEncrypt && !*isDecrypt {
		usageAndExit("Missing encrypt/decrypt")
	}

	var err error

	filename, err := filepath.Abs(*inputFile)
	if err != nil {
		errorAndExit(err)
	}

	if *isEncrypt {
		err = EncryptFile(*key, filename)
	} else if *isDecrypt {
		err = DecryptFile(*key, filename)
	}

	if err != nil {
		errorAndExit(err)
	}
}
