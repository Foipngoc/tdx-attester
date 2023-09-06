// tdx_attester project main.go
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {

	http.HandleFunc("/attest", attest)
	log.Println("server监听于端口:9091")
	h1 := http.FileServer(http.Dir("files"))
	http.Handle("/", h1)
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func attest(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("quote")
	if err != nil {
		log.Println("Error retrieving the file from the form data.")
		return
	}
	defer file.Close()

	pubk := r.FormValue("pubk")
	log.Println(pubk)
	// 创建目标文件
	output, err := os.Create("quote.dat")
	if err != nil {
		log.Println("Error creating the output file.")
		return
	}
	defer output.Close()

	// 将上传的文件内容复制到目标文件
	_, err = io.Copy(output, file)
	if err != nil {
		log.Println("Error copying the file.")
		return
	}

	cmd := exec.Command("app", "-quote", "quote.dat")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("错误:\n%s\n", string(out))
		fmt.Printf("认证失败：%s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	flag := "Error"
	if strings.Contains(string(out), flag) {
		fmt.Printf("认证失败:\n%s\n", string(out))
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)

	} else {
		fmt.Printf("认证成功:\n%s\n", string(out))

		publicKeyBlock, _ := pem.Decode([]byte(pubk))
		if publicKeyBlock == nil || publicKeyBlock.Type != "RSA PUBLIC KEY" {
			panic("无法解码公钥")
		}
		publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
		if err != nil {
			panic(err)
		}
		/*rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
		if !ok {
			panic("无法将公钥转换为RSA类型")
		}*/

		encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte("5ac3b50421eeb26f55141a537cb322ff"), nil)
		fmt.Printf("配置key.\n")

		w.WriteHeader(http.StatusOK)

		w.Write(encryptedBytes)
	}

}
