package main

import (
	"io"
	"net/http"

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
func HandleHello(c *gin.Context) {
	who := c.Query("who")
	c.JSON(http.StatusOK, gin.H{"msg": "hello " + who})
}

// HandleLogin doc
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

	f := &File{Name: fileHeader.Filename, Len: int(fileHeader.Size), Content: fileCon[:n]}
	files.Files = append(files.Files, f)
	files.Len++
	c.JSON(http.StatusOK, f)
}

var files = Files{Files: []*File{}}

// File doc
type File struct {
	Name    string `json:"name"`
	Len     int    `json:"len"`
	Content []byte `json:"content"`
}

// Files doc
type Files struct {
	Files []*File `json:"files"`
	Len   int     `json:"len"`
}

// HandleList doc
func HandleList(c *gin.Context) {
	c.JSON(http.StatusOK, files)
}
