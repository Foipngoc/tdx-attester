// tdx_attester project main.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	http.HandleFunc("/attest", attest)
	log.Println("server监听于端口:4439")
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
		log.Fatalf("认证失败：%s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("key"))

}
