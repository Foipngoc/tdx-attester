package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var (
	privateKey *rsa.PrivateKey
	pubk       []byte
)

func genRSA() {
	var err error
	// 生成RSA密钥对
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("无法生成私钥:", err)
		return
	}

	/*
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		}
			// 将私钥保存到文件
			privateKeyFile, err := os.Create("private_key.pem")
			if err != nil {
				fmt.Println("无法创建私钥文件:", err)
				return
			}
			defer privateKeyFile.Close()
			if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
				fmt.Println("无法写入私钥文件:", err)
				return
			}
			prvk = pem.EncodeToMemory(privateKeyPEM)*/

	// 从私钥获取公钥
	publicKey := privateKey.PublicKey
	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
	}
	/*
		// 将公钥保存到文件
		publicKeyFile, err := os.Create("public_key.pem")
		if err != nil {
			fmt.Println("无法创建公钥文件:", err)
			return
		}
		defer publicKeyFile.Close()

		if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
			fmt.Println("无法写入公钥文件:", err)
			return
		}*/
	pubk = pem.EncodeToMemory(publicKeyPEM)

}

//注意client 本身是连接池，不要每次请求时创建client
var (
	HttpClient = &http.Client{
		Timeout: 3 * time.Second,
	}
)

// 上传文件
// url                请求地址
// params        post form里数据
// nameField  请求地址上传文件对应field
// fileName     文件名
// file               文件
func UploadFile(url string, params map[string]string, nameField, fileName string, file io.Reader) {
	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	formFile, err := writer.CreateFormFile(nameField, fileName)
	if err != nil {
		return
	}

	_, err = io.Copy(formFile, file)
	if err != nil {
		return
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}
	//req.Header.Set("Content-Type","multipart/form-data")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := HttpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println("认证失败：", string(content))
		return
	}

	//decrypt xxx
	decryptedBytes, err := privateKey.Decrypt(nil, content, &rsa.OAEPOptions{Hash: crypto.SHA256})
	fmt.Println("获得key:", string(decryptedBytes))
	//launch app
	cmd := exec.Command("python3", "pytorchexample.py")
	cmd.Dir = "/root/examples/pytorch/"
	out, _ := cmd.CombinedOutput()
	fmt.Println("执行torch：", string(out))

}
func main() {
	genRSA()
	params := make(map[string]string)
	params["pubk"] = string(pubk)

	file, err := os.Open("/opt/intel/tdx-quote-generation-sample/quote.dat")
	if err != nil {
		fmt.Println("打开文件时发生错误:", err)
		return
	}

	UploadFile("http://192.168.122.1:9091/attest", params, "quote", "quote.dat", file)

}
