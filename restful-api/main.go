package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	store := sessions.NewCookieStore([]byte("dashdjkasdhaksda"))
	store.Options(sessions.Options{
		MaxAge:   7200,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("mydemo", store))
	r.NoRoute(func(c *gin.Context) { c.JSON(http.StatusNotFound, "Invaild api request") })

	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", HandleLogin)
		v1.GET("/articles", HandleGetArticles)

		v1Auth := v1.Use(Auth)
		{
			v1Auth.POST("/logout", HandleLogout)
			v1Auth.POST("/articles", HandlePostArticles)
			v1Auth.PUT("/articles", HandleUpdateArticles)
			v1Auth.DELETE("/articles/:id", HandleDeleteArticles)
			v1.GET("/articles/:id/comments", HandleGetComments)
			v1Auth.POST("/articles/:id/comments", HandleAddComments)
			v1Auth.PUT("/articles/:id/comments/:id", HandleUpdateComments)
			v1Auth.DELETE("/articles/:id/comments/:id", HandleDeleteComments)
		}
	}

	r.Run(":8080")
}

const sessionsKey = "user"

// Auth doc
func Auth(c *gin.Context) {
	session := sessions.Default(c)
	u := session.Get(sessionsKey)
	if u == nil {
		c.JSON(http.StatusUnauthorized, &Resp{Error: "please login"})
		c.Abort()
		return
	}
}

// Resp doc
type Resp struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

// LoginParams doc
type LoginParams struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// HandleLogin doc
func HandleLogin(c *gin.Context) {
	param := &LoginParams{}
	if err := c.BindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
		return
	}

	// 做一些校验
	// ...

	session := sessions.Default(c)
	session.Set(sessionsKey, param.UserID)
	session.Save()
	c.JSON(http.StatusOK, &Resp{Data: "login succeed"})
}

// 模拟数据库
var tempStorage = []*Article{}

// HandleLogout doc
func HandleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(sessionsKey)
	session.Save()
	c.JSON(http.StatusOK, &Resp{Data: "logout succeed"})
}

// HandleGetArticles doc
func HandleGetArticles(c *gin.Context) {

	page := c.Query("page")
	pageSize := c.Query("page_size")
	orderby := c.Query("order")
	searchKey := c.Query("search")

	// 分页
	// 查询
	// 排序
	// ...
	_, _, _, _ = page, pageSize, orderby, searchKey

	c.JSON(http.StatusOK, &Resp{Data: map[string]interface{}{
		"result": tempStorage,
		"total":  len(tempStorage),
	}})
}

// Article doc
type Article struct {
	ID       int       `json:"id"`
	Titile   string    `json:"titile"`
	Tags     []string  `json:"tags"`
	Content  string    `json:"content"`
	UpdateAt time.Time `json:"update_at"`
	CreateAt time.Time `json:"create_at"`
}

// HandlePostArticles doc
func HandlePostArticles(c *gin.Context) {
	param := &Article{}
	if err := c.BindJSON(param); err == nil {
		c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
		return
	}

	// 参数判断
	// 保存文章
	// ...
	param.ID = len(tempStorage)
	param.CreateAt = time.Now()
	tempStorage = append(tempStorage, param)

	c.JSON(http.StatusOK, &Resp{Data: param})
}

// HandleUpdateArticles doc
func HandleUpdateArticles(c *gin.Context) {
	param := &Article{}
	if err := c.BindJSON(param); err == nil {
		c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
		return
	}

	// 参数判断
	// 保存文章
	// ...
	param.UpdateAt = time.Now()
	for i, v := range tempStorage {
		if v.ID == param.ID {
			param.CreateAt = v.CreateAt
			tempStorage[i] = param
			break
		}
	}

	c.JSON(http.StatusOK, &Resp{Data: param})
}

// HandleDeleteArticles doc
func HandleDeleteArticles(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
		return
	}

	// 删除
	// ...
	for i, v := range tempStorage {
		if v.ID == id {
			tempStorage = append(tempStorage[:i], tempStorage[i+1:]...)
			break
		}
	}

	c.JSON(http.StatusOK, &Resp{Data: "delete succeed"})
}

// HandleGetComments doc
func HandleGetComments(c *gin.Context) {

}

// HandleAddComments doc
func HandleAddComments(c *gin.Context) {

}

// HandleUpdateComments doc
func HandleUpdateComments(c *gin.Context) {

}

// HandleDeleteComments doc
func HandleDeleteComments(c *gin.Context) {

}
