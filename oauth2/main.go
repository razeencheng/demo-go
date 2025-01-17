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

type GithubUserInfo struct {
	Name string `json:"name"`
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

	// 通过github返回的code码，再向github获取access token，只有使用access token才能获取用户资源
	accessToken, err := getAccessTokenByCode(code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 通过access token获取用户资源，这里的资源为用户的名称
	username, err := getUsername(accessToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 最后获取到用户信息后，我们重定向到欢迎页面，也就是表示用户登录成功
	w.Header().Set("Location", "/welcome.html?username="+username)
	w.WriteHeader(http.StatusFound)
}

func getAccessTokenByCode(code string) (string, error) {
	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		clientID, clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		log.Printf("could not create HTTP request: %v", err)
		return "", err
	}

	// 设置我们期待返回的格式为json
	req.Header.Set("accept", "application/json")

	// 发送http请求
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("could not send HTTP request: %v", err)
		return "", err
	}
	defer res.Body.Close()

	// 解析
	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		log.Printf("could not parse JSON response: %v", err)
		return "", err
	}

	return t.AccessToken, nil
}

func getUsername(accessToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		log.Printf("could not create HTTP request: %v", err)
		return "", err
	}

	// 设置我们期待返回的格式为json
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 发送http请求
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("could not send HTTP request: %v", err)
		return "", err
	}
	defer res.Body.Close()

	// 解析
	var u GithubUserInfo
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		log.Printf("could not parse JSON response: %v", err)
		return "", err
	}

	return u.Name, nil
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
