package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

// 你在注册时得到的
const (
	clientID     = "你的客户端ID"
	clientSecret = "你的客户端密钥"
)

var httpClient = http.Client{}

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/oauth/redirect", HandleOAuthRedirect)

	http.ListenAndServe(":8080", nil)
}

// HandleOAuthRedirect doc
func HandleOAuthRedirect(w http.ResponseWriter, r *http.Request) {
	// 首先，我们从URI中解析出code参数
	// 如: http://localhost:8080/oauth/redirect?code=260f17a7308f2c566725
	err := r.ParseForm()
	if err != nil {
		log.Printf("could not parse query: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	code := r.FormValue("code")

	// 接下来，我们通过 clientID,clientSecret,code 获取授权密钥
	// 前者是我们在注册时得到的，后者是用户确认后，重定向到该路由，从中获取到的。
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientID, clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		log.Printf("could not create HTTP request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// 设置我们期待返回的格式为json
	req.Header.Set("accept", "application/json")

	// 发送http请求
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("could not send HTTP request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer res.Body.Close()

	// 解析
	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		log.Printf("could not parse JSON response: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// 最后获取到access_token后，我们重定向到欢迎页面，也就是表示用户登录成功，同属获取一些用户的基本展示信息
	w.Header().Set("Location", "/welcome.html?access_token="+t.AccessToken)
	w.WriteHeader(http.StatusFound)
}

func init() {
	tmpl, err := template.ParseFiles("public/index.tmpl")
	if err != nil {
		log.Fatalf("parse html templ err: %v", err)
	}

	file, err := os.OpenFile("public/index.html", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0660)
	if err != nil {
		log.Fatalf("create index.html err: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, map[string]interface{}{
		"ClientId": clientID,
	})
	if err != nil {
		log.Fatalf("exec tmpl err: %v", err)
	}
}
