
# [Go学习笔记(七) | 理解并实现 OAuth 2.0](https://razeen.me/post/oauth2-protocol-details.html)

OAuth 2.0是一个关于授权的开放网络标准，主要致力于简化客户端人员开发，同时为Web应用程序、桌面应用程序、移动电话和物联网设备提供特定的授权规范。他的官网在[这里](https://oauth.net/2/)。在[RFC6749](https://tools.ietf.org/html/rfc6749)中有明确协议规范。

简单来说，我们平时使用的很多第三方登录并获取头像等信息就是用的OAuth 2.0。如我们用QQ登录一些论坛，用google账号登陆facebook，用github账号登陆gitlab等。如下图展示的就是利用QQ登录网易云音乐Web版，其中用到的就是OAuth 2.0。

<!--more-->



![](https://st.razeen.me/image/blog/oauth2-01.png)



### 理解OAuth 2.0

在开始编程前，我们先要掌握OAuth 2.0认证的流程是什么样的，下面我们一起走进OAuth 2.0。



***下面的情景都以我用QQ登录网易云Web版为例子。***



#### 相关名词

在理解OAuth 2.0前，我们先需要理解下面这几个名词。

- `resource owner`： 资源所有者。也就是例子中的我(用户)。

- `resource server`： 资源服务器，即服务提供商存放用户资源的服务器。例子中就是QQ的资源服务器，存放了我的QQ昵称，QQ头像，性别等等信息。

- ` client`： 客户端，需要得到资源的应用程序。例子中的网易云音乐Web版。

- `authorization server`：认证服务器，即服务提供商专门用来处理认证的服务器。例子中QQ提供的认证服务器(跳出来让我们扫码登录QQ的这个服务)。

- `user-agent`：用户代理。我们用来访问客户端的程序，如这里就是浏览器。



总体说来，这种登录流程大概就是这样：

我`(resource owner)`需要用QQ`(resource server)` 在浏览器`(user-agent)`登录网易云音乐`(client)`，我们先登录QQ`(authorization server)`，授权给网易云。然后网易云才能拿着这个授权信息去QQ资源服务器`(resource server)`获取到我的头像/昵称/性别等信息。

具体流程我们继续往下看。



#### 协议流程

下图来自[RFC6749](https://tools.ietf.org/html/rfc6749)。

```txt
     +--------+                               +---------------+
     |        |--(A)- Authorization Request ->|   Resource    |
     |        |                               |     Owner     |
     |        |<-(B)-- Authorization Grant ---|               |
     |        |                               +---------------+
     |        |
     |        |                               +---------------+
     |        |--(C)-- Authorization Grant -->| Authorization |
     | Client |                               |     Server    |
     |        |<-(D)----- Access Token -------|               |
     |        |                               +---------------+
     |        |
     |        |                               +---------------+
     |        |--(E)----- Access Token ------>|    Resource   |
     |        |                               |     Server    |
     |        |<-(F)--- Protected Resource ---|               |
     +--------+                               +---------------+

                     Figure 1: Abstract Protocol Flow
```

- (A) 客户端请求`resource owner`授权；

  这种授权可以是直接向`resource owner`请求，也可以通过`authorization server`间接请求。例子中，就是跳转到QQ的`authorization server`，让用户登录QQ。

- (B) 用户同意给予授权；

  例子中，用户点击同意，重定向回网易云音乐，这是一种授权模式。在RFC中规定了4中授权模式，具体采用哪一种取决于`authorization server`，下一小节具体介绍。

- (C) 客户端拿着上一步的授权，向`authorization server`申请令牌(access_token)；

- (D) `authorization server`确认授权无误后，发放令牌(access_token)；

- (E) 客户端拿着令牌（access_token）到`resource server`去获取资源;

  例子中，获取QQ头像，昵称等。

- (F) `resource server`确认无误，同意向客户端下发受保护的资源。



下面详细说到其中四种授权模式。



#### 授权模式

客户端要得到令牌(access_token), 必须需要得到用户的授权，在OAuth 2.0 中定义了四种授权模式：

- 授权码模式 (Authorization Code)
- 简化模式 (Implicit)
- 密码模式 (Resource Owner Password Credentials)
- 客户端模式 (Client Credentials)

每种模式的使用场景与流程都有一定的差别。



##### 授权码模式 (Authorization Code)

授权码模式的授权流程是基于重定向。我们例子中的用QQ登录网易云就是这种模式，流程图如下(来自RFC6749)：

```
    +----------+
     | Resource |
     |   Owner  |
     |          |
     +----------+
          ^
          |
         (B)
     +----|-----+          Client Identifier      +---------------+
     |         -+----(A)-- & Redirection URI ---->|               |
     |  User-   |                                 | Authorization |
     |  Agent  -+----(B)-- User authenticates --->|     Server    |
     |          |                                 |               |
     |         -+----(C)-- Authorization Code ---<|               |
     +-|----|---+                                 +---------------+
       |    |                                         ^      v
      (A)  (C)                                        |      |
       |    |                                         |      |
       ^    v                                         |      |
     +---------+                                      |      |
     |         |>---(D)-- Authorization Code ---------'      |
     |  Client |          & Redirection URI                  |
     |         |                                             |
     |         |<---(E)----- Access Token -------------------'
     +---------+       (w/ Optional Refresh Token)

   Note: The lines illustrating steps (A), (B), and (C) are broken into
   two parts as they pass through the user-agent.

                     Figure 3: Authorization Code Flow
```

- (A) 用户访问客户端，客户端将用户重定向到认证服务器；
- (B) 用户选择是否授权；
- (C) 如果用户同意授权，认证服务器重定向到客户端事先指定的地址，而且带上授权码(code)；
- (D) 客户端收到授权码，带着前面的重定向地址，向认证服务器申请访问令牌；
- (E) 认证服务器核对授权码与重定向地址，确认后向客户端发送访问令牌和更新令牌(可选)。



1）在A中，客户端申请授权，重定向到认证服务器的URI中需要包含这些参数：

| 参数名称      | 参数含义                                                     | 是否必须 |
| ------------- | ------------------------------------------------------------ | -------- |
| response_type | 授权类型，此处的值为`code`                                   | 必须     |
| client_id     | 客户端ID，客户端到资源服务器注册的ID                         | 必须     |
| redirect_uri  | 重定向URI                                                    | 可选     |
| scope         | 申请的权限范围，多个逗号隔开                                 | 可选     |
| state         | 客户端的当前状态，可以指定任意值，认证服务器会原封不动的返回这个值 | 推荐     |

RFC6749中例子如下：

```bash
   GET /authorize?response_type=code&client_id=s6BhdRkqt3&state=xyz
        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
    Host: server.example.com
```

我们在用QQ登录网易云音乐时实际的地址如下：

```bash
https://graph.qq.com/oauth2.0/show?which=Login
		&display=pc
		&client_id=100495085
		&response_type=code
		&redirect_uri=https://music.163.com/back/qq
		&forcelogin=true
		&state=ichQQlAgNi
		&checkToken=sretfyguihojpr
```

我们看到参数基本一致，至于还多几个，有什么用这里就没有去深究了。



2）在C中，认证服务器返回的URI中，需要包含下面这些参数：

| 参数名称 | 参数含义                                                     | 是否必须 |
| -------- | ------------------------------------------------------------ | -------- |
| code     | 授权码。认证服务器返回的授权码，生命周期不超过10分钟，而且要求只能使用一次，和A中的`client_id`,`redirect_uri`绑定。 | 必须     |
| state    | 如果A中请求包含这个参数，资源服务器原封不动的返回。          | 可选     |

如：

```bash
     HTTP/1.1 302 Found
     Location: https://client.example.com/cb?code=SplxlOBeZQQYbYS6WxSbIA
               &state=xyz
```



3) 在D中客户端向认证服务器申请令牌(access_token)时，需要包含下面这些参数。

| 参名称       | 参数含义                               | 是否必须 |
| ------------ | -------------------------------------- | -------- |
| grant_type   | 授权模式，此处为`authorization_code`。 | 必须     |
| code         | 授权码，C中获取的`code`。              | 必须     |
| redirect_uri | 重定向URI，需要和A中一致。             | 必须     |
| client_id    | 客户端ID，与A中一致。                  | 必须     |

如：

```bash
     POST /token HTTP/1.1
     Host: server.example.com
     Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
     Content-Type: application/x-www-form-urlencoded

     grant_type=authorization_code&code=SplxlOBeZQQYbYS6WxSbIA
     &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb
```



4) 在E中，认证服务器返回的信息中，包含下面参数：

| 参数名称      | 参数含义                                                   | 是否必须 |
| ------------- | ---------------------------------------------------------- | -------- |
| access_token  | 访问令牌                                                   | 必须     |
| token_type    | 令牌类型，大小写不敏感。例如 Bearer，MAC。                 | 必须     |
| expires_in    | 过期时间(s)， 如果不设置也要通过其他方法设置一个。         | 推荐     |
| refresh_token | 更新令牌的token。当令牌过期的时候，可用通过该值刷新token。 | 可选     |
| scope         | 权限范围，如果与客户端申请范围一致，可省略。               | 可选     |

如：

```bash
     HTTP/1.1 200 OK
     Content-Type: application/json;charset=UTF-8
     Cache-Control: no-store
     Pragma: no-cache

     {
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "expires_in":3600,
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }
```



5) 如果我们的令牌过期了，需要更新，这里就需要使用`refresh_token`获取一个新令牌了。此时发起HTTP请求需要的参数有：

| 参数名称      | 参数含义                        | 是否必须 |
| ------------- | ------------------------------- | -------- |
| grant_type    | 授权类型，此处是`refresh_token` | 必须     |
| refresh_token | 更新令牌的token。               | 必须     |
| scope         | 权限范围。                      | 可选     |

如：

```bash
     POST /token HTTP/1.1
     Host: server.example.com
     Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
     Content-Type: application/x-www-form-urlencoded

     grant_type=refresh_token&refresh_token=tGzv3JOkF0XG5Qx2TlKWIA
```



这就是授权码模式，应该也是我们平常见的较多的模式了。



##### 简化模式 (Implicit)

简化模式，相当于授权码模式中，C步骤不再通过客户端，直接在浏览器`(user-agent)`中向认证服务器申请令牌，认证服务器不再返回授权码，所有步骤都在浏览器中完成，最后资源服务器将令牌放在`Fragment`中，浏览器从中将令牌提取，发送给客户端。

所以这个令牌对访问者时可见的，而且客户端不需要认证。详细流程如下。

```txt
     +----------+
     | Resource |
     |  Owner   |
     |          |
     +----------+
          ^
          |
         (B)
     +----|-----+          Client Identifier     +---------------+
     |         -+----(A)-- & Redirection URI --->|               |
     |  User-   |                                | Authorization |
     |  Agent  -|----(B)-- User authenticates -->|     Server    |
     |          |                                |               |
     |          |<---(C)--- Redirection URI ----<|               |
     |          |          with Access Token     +---------------+
     |          |            in Fragment
     |          |                                +---------------+
     |          |----(D)--- Redirection URI ---->|   Web-Hosted  |
     |          |          without Fragment      |     Client    |
     |          |                                |    Resource   |
     |     (F)  |<---(E)------- Script ---------<|               |
     |          |                                +---------------+
     +-|--------+
       |    |
      (A)  (G) Access Token
       |    |
       ^    v
     +---------+
     |         |
     |  Client |
     |         |
     +---------+

```

- (A) 客户端将用户导向认证服务器， 携带客户端ID及重定向URI；
- (B) 用户是否授权；
- (C) 用户同意授权后，认证服务器重定向到A中指定的URI，并且在URI的`Fragment`中包含了访问令牌；
- (D) 浏览器向资源服务器发出请求，该请求中不包含C中的`Fragment`值；
- (E) 资源服务器返回一个网页，其中包含了可以提取C中`Fragment`里面访问令牌的脚本；
- (F) 浏览器执行E中获得的脚本，提取令牌；
- (G) 浏览器将令牌发送给客户端。



1）在A步骤中，客户端发送请求，需要包含这些参数：

| 参数名称      | 参数含义                                       | 是否必须 |
| ------------- | ---------------------------------------------- | -------- |
| response_type | 授权类型，此处值为`token`。                    | 必须     |
| client_id     | 客户端的ID。                                   | 必须     |
| redirect_uri  | 重定向的URI。                                  | 可选     |
| scope         | 权限范围。                                     | 可选     |
| state         | 客户端的当前状态。指定后服务器会原封不动返回。 | 推荐     |

如：

```bash
    GET /authorize?response_type=token&client_id=s6BhdRkqt3&state=xyz
        &redirect_uri=https%3A%2F%2Fclient%2Eexample%2Ecom%2Fcb HTTP/1.1
    Host: server.example.com
```



2) 在C中，认证服务器返回的URI中，参数主要有：

| 参数名称     | 参数含义                               | 是否必须 |
| ------------ | -------------------------------------- | -------- |
| access_token | 访问令牌。                             | 必须     |
| token_type   | 令牌类型。                             | 必须     |
| expires_in   | 过期时间。                             | 推荐     |
| scope        | 权限范围。                             | 可选     |
| state        | 客户端访问时如果指定了，原封不动返回。 | 可选     |

如：

```bash
     HTTP/1.1 302 Found
     Location: http://example.com/cb#access_token=2YotnFZFEjr1zCsicMWpAA
               &state=xyz&token_type=example&expires_in=3600
```

我们可以看到C中返回的是一个重定向，而重定向的这个网址的`Fragment`部分包含了令牌。

D步骤中就是访问这个重定向指定的URI，而且不带`Fragment`部分，服务器会返回从`Fragment`中提取令牌的脚本，最后浏览器运行脚本获取到令牌发送给客户端。



##### 密码模式 (Resource Owner Password Credentials)

密码模式就是用户直接将用户名密码提供给客户端，客户端使用这些信息到认证服务器请求授权。具体流程如下：

```
     +----------+
     | Resource |
     |  Owner   |
     |          |
     +----------+
          v
          |    Resource Owner
         (A) Password Credentials
          |
          v
     +---------+                                  +---------------+
     |         |>--(B)---- Resource Owner ------->|               |
     |         |         Password Credentials     | Authorization |
     | Client  |                                  |     Server    |
     |         |<--(C)---- Access Token ---------<|               |
     |         |    (w/ Optional Refresh Token)   |               |
     +---------+                                  +---------------+

            Figure 5: Resource Owner Password Credentials Flow
```

- (A) 资源所有者提供用户名密码给客户端；
- (B) 客户端拿着用户名密码去认证服务器请求令牌；
- (C) 认证服务器确认后，返回令牌；



1) 在B中客户端发送的请求中，需要包含这些参数：

| 参数名称   | 参数含义                       | 是否必须 |
| ---------- | ------------------------------ | -------- |
| grant_type | 授权类型，此处值为`password`。 | 必须     |
| username   | 用户名。                       | 必须     |
| password   | 用户的密码。                   | 必须     |
| scope      | 权限范围。                     | 可选     |

如：

```bash
     POST /token HTTP/1.1
     Host: server.example.com
     Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
     Content-Type: application/x-www-form-urlencoded

     grant_type=password&username=johndoe&password=A3ddj3w
```



3) 在C中，认证服务器返回访问令牌。如：

```bash
     HTTP/1.1 200 OK
     Content-Type: application/json;charset=UTF-8
     Cache-Control: no-store
     Pragma: no-cache

     {
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "expires_in":3600,
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }
```



#####  客户端模式 (Client Credentials)

客户端模式，其实就是客户端直接向认证服务器请求令牌。而用户直接在客户端注册即可，一般用于后端 API 的相关操作。其流程如下：

```

     +---------+                                  +---------------+
     |         |                                  |               |
     |         |>--(A)- Client Authentication --->| Authorization |
     | Client  |                                  |     Server    |
     |         |<--(B)---- Access Token ---------<|               |
     |         |                                  |               |
     +---------+                                  +---------------+

                     Figure 6: Client Credentials Flow
```

- (A) 客户端发起身份认证，请求访问令牌；
- (B) 认证服务器确认无误，返回访问令牌。



1) 在A中，客户端发起请求的参数有：

| 参数名称   | 参数含义                                 | 是否必须 |
| ---------- | ---------------------------------------- | -------- |
| grant_type | 授权类型，此处值为`client_credentials`。 | 必须     |
| scope      | 权限范围。                               | 可选     |

如：

``` bash
     POST /token HTTP/1.1
     Host: server.example.com
     Authorization: Basic czZCaGRSa3F0MzpnWDFmQmF0M2JW
     Content-Type: application/x-www-form-urlencoded

     grant_type=client_credentials
```



2） 认证服务器认证后，发放访问令牌，如：

```bash
     HTTP/1.1 200 OK
     Content-Type: application/json;charset=UTF-8
     Cache-Control: no-store
     Pragma: no-cache

     {
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "expires_in":3600,
       "example_parameter":"example_value"
     }
```





### 实现OAuth 2.0

上面我们一起了解了OAuth 2.0的流程，现在我们开始实现OAuth 2.0。

首先，我需要一个使用场景，大概是这样的。

> 我有一个需要登录的应用（web app)。
> 用户可以选择使用github登录。
> 登录后，显示一个简单的欢迎页面。

有了这个场景后，我们开始设计（这里我们实现的是第一种授权码模式）。



#### 登录页面

在这里，我们充当的角色就是`client`,我们需要一个简单的页面，让用户选择使用`github`登录。

这是一个简单的html页面，`public/index.html`页面。

```html
<!DOCTYPE html>
<html>

<body>
  <a href="https://github.com/login/oauth/authorize?client_id=89ac6f58f15f658d8dd5&redirect_uri=http://localhost:8080/oauth/redirect">
    Login with github
  </a>
</body>

</html>
```

也就是当用户点击`Login with github`会访问

```bash
https://github.com/login/oauth/authorize?
		client_id=89ac6f58f15f658d8dd5
		&redirect_uri=http://localhost:8080/oauth/redirect
```

其中：

- `https://github.com/login/oauth/authorize` 是GitHub的OAuth网关地址。

- `client_id=89ac6f58f15f658d8dd5` 这个是我申请的客户端ID，

  需要到<https://github.com/settings/applications/new>注册。

  注册过程很简单，但要注意，最后一栏`Authorization callback URL`是和下面`redirect_uri`一致。

  如：

  ![](https://st.razeen.me/image/blog/Jietu20190208-230400.png)

  成功后你会获的`Client ID`与`Client Secret`, 前者就是这里的`client_id`，两者在后面的编码中需要用到。

- `redirect_uri=http://localhost:8080/oauth/redirect` 当用户确认，获取权限后重定向到的该地址。



然后我们写个`mian.go`, 启动一个简单的服务。

```go
func main() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.ListenAndServe(":8080", nil)
}
```

写到这里，我们代码已经能跑起了，如果你点击`Login with github`我们会看到如下的授权页面。

![](https://st.razeen.me/image/blog/Jietu20190208-232606.png)
当我们确定过后，会重定向到我们指定的地址，而且带上`code`,如：

```
http://localhost:8080/oauth/redirect?code=260f17a7308f2c566725
```

此时我们需要做的是，拿着该code去请求访问令牌(access_token)，对应授权码模式中的D步骤。



#### 重定向路由

现在我们有`code`了，我们需要去请求令牌。接下来我们就需要向`https://github.com/login/oauth/access_token`发送POST请求获取访问令牌(access_token)。

> 关于更多GitHub重定向URI的信息你可以看[这里](https://developer.github.com/apps/building-oauth-apps/authorizing-oauth-apps/#2-users-are-redirected-back-to-your-site-by-github)。

现在，让我们将`/oauth/redirect`路由补全，在该步骤中请求访问令牌(access_token)。

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
```



#### 欢迎页面

上面我们拿到了访问令牌，同时，我们重定向到了欢迎页面。接下来，我们就在欢迎页中获取用户的GitHub名称作为展示。

添加`public/welcome.html`页面。

```html
<!DOCTYPE html>
<html lang="en">

<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="ie=edge">
	<title>Hello</title>
</head>

<body>

</body>
<script>
	// 获取access_token
	const query = window.location.search.substring(1)
	const token = query.split('access_token=')[1]

	// 访问资源服务器地址，获取相关资源
	fetch('https://api.github.com/user', {
			headers: {
                // 将token放在Header中
				Authorization: 'token ' + token
			}
		})
		// 解析返回的JSON
		.then(res => res.json())
		.then(res => {
            // 这里我们能得到很多信息
			// 具体看这里 https://developer.github.com/v3/users/#get-the-authenticated-user
			// 这里我们就只展示一下用户名了
			const nameNode = document.createTextNode(`Welcome, ${res.name}`)
			document.body.appendChild(nameNode)
		})
</script>
```

到这里，我们再次运行程序，点击`Login with github`，同意后我们可以看到类似`Welcome, Razeen`的信息，此时我们已经完成整个OAuth 2.0的流程了。

> 获得授权之后，我们能访问的API不仅仅有这些，更多的请看[这里](https://developer.github.com/v3/)。



*源码在[这里](https://github.com/razeencheng/demo-go/tree/master/oauth2)。*



### 关于安全

1. 这里我们将访问令牌直接放在了URI中，这么做其实是不安全的。更好的做法是我们创建一个会话session，将cooike发送给用户即可。

2. 在前面OAuth 2.0的解释中，我们知道在请求权限的过程中，我们可以加上`state`字段，如

   ```bash
   https://github.com/login/oauth/authorize？
   		client_id=89ac6f58f15f658d8dd5
   		&redirect_uri=http://localhost:8080/oauth/redirect
   		&state=xxxxx
   ```

   而服务器会原封不动的返回该state。这样我们可以将该字段设置一些随机值，如果资源服务器返回的`state`值与我们设置的不同，我们认为该请求不是来自正确的资源服务器，应该拒绝。





**参考**

- [FRC6749](https://tools.ietf.org/html/rfc6749)

- [Implementing OAuth 2.0 with Go(Golang)](https://www.sohamkamani.com/blog/golang/2018-06-24-oauth-with-golang/)

- [10 分钟理解什么是 OAuth 2.0 协议](https://deepzz.com/post/what-is-oauth2-protocol.html)

- [理解OAuth 2.0](http://www.ruanyifeng.com/blog/2014/05/oauth_2_0.html)
