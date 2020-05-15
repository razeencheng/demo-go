package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

const isLoginKey = "is_login"

// HandleAuth doc
func HandleAuth(c *gin.Context) {
	session := sessions.Default(c)
	isLogin := session.Get(isLoginKey)
	if isLogin == nil || !isLogin.(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "please login"})
		c.Abort()
		return
	}
}

// HandleHello doc
// @Summary 测试SayHello
// @Description 向你说Hello
// @Tags 测试
// @Accept mpfd
// @Produce json
// @Param who query string true "人名"
// @Success 200 {string} string "{"msg": "hello Razeen"}"
// @Failure 400 {string} string "{"msg": "who are you"}"
// @Router /hello [get]
func HandleHello(c *gin.Context) {
	who := c.Query("who")

	if who == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "who are u?"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "hello " + who})
}

// HandleLogin doc
// @Summary 登陆
// @Tags 登陆
// @Description 登入
// @Accept mpfd
// @Produce json
// @Param user formData string true "用户名" default(admin)
// @Param password formData string true "密码"
// @Success 200 {string} string "{"msg":"login success"}"
// @Failure 400 {string} string "{"msg": "user or password error"}"
// @Router /login [post]
func HandleLogin(c *gin.Context) {
	user := c.PostForm("user")
	pwd := c.PostForm("password")

	if user == "admin" && pwd == "123456" {
		session := sessions.Default(c)
		session.Set(isLoginKey, true)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"msg": "login success"})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"msg": "user or password error"})
}

// HandleUpload doc
// @Summary 上传文件
// @Tags 文件处理
// @Description 上传文件
// @Accept mpfd
// @Produce json
// @Param file formData file true "文件"
// @Success 200 {object} main.File
// @Router /upload [post]
func HandleUpload(c *gin.Context) {

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}

	fileCon := make([]byte, 1<<20)
	n, err := file.Read(fileCon)
	if err != nil {
		if err != io.EOF {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}
	}

	id++
	f := &File{ID: id, Name: fileHeader.Filename, Len: int(fileHeader.Size), Content: fileCon[:n]}
	files.Files = append(files.Files, f)
	files.Len++
	c.JSON(http.StatusOK, f)
}

var files = Files{Files: []*File{}}
var id int

// File doc
type File struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Len     int    `json:"len"`
	Content []byte `json:"-"`
}

// Files doc
type Files struct {
	Files []*File `json:"files"`
	Len   int     `json:"len"`
}

// HandleList doc
// @Summary 查看文件列表
// @Tags 文件处理
// @Description 文件列表
// @Accept mpfd
// @Produce json
// @Success 200 {object} main.Files
// @Router /list [get]
func HandleList(c *gin.Context) {
	c.JSON(http.StatusOK, files)
}

// HandleGetFile doc
// @Summary 获取某个文件
// @Tags 文件处理
// @Description 获取文件
// @Accept mpfd
// @Produce octet-stream
// @Param id path integer true "文件ID"
// @Success 200 {string} string ""
// @Router /file/{id} [get]
func HandleGetFile(c *gin.Context) {
	fid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}

	for _, f := range files.Files {
		if f.ID == fid {
			c.Writer.WriteHeader(http.StatusOK)
			c.Header("Access-Control-Expose-Headers", "Content-Disposition")
			c.Header("Content-Disposition", "attachment; "+f.Name)
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Accept-Length", fmt.Sprintf("%d", len(f.Content)))
			c.Writer.Write(f.Content)
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"msg": "no avail file"})
}

// JSONParams doc
type JSONParams struct {
	// 这是一个字符串
	Str string `json:"str"`
	// 这是一个数字
	Int int `json:"int"`
	// 这是一个字符串数组
	Array []string `json:"array"`
	// 这是一个结构
	Struct struct {
		Field string `json:"field"`
	} `json:"struct"`
}

// HandleJSON doc
// @Summary 获取JSON的示例
// @Tags JSON
// @Description 获取JSON的示例
// @Accept json
// @Produce json
// @Param param body main.JSONParams true "需要上传的JSON"
// @Success 200 {object} main.JSONParams "返回"
// @Router /json [post]
func HandleJSON(c *gin.Context) {
	param := JSONParams{}
	if err := c.BindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, param)
}
