package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

type FrameworkController struct {
	baseDir string
}

func NewFrameworkController() Router {
	f := FrameworkController{baseDir: os.Getenv("BASE_DIR")}
	if f.baseDir == "" {
		f.baseDir = "."
	}
	return &f
}

func (i *FrameworkController) Handler(c *gin.Engine) {
	c.GET("/framework/:framework", i.getVersions)
	c.GET("/framework/:framework/:version", i.getFileInfo)
	c.GET("/framework/:framework/:version/*filepath", i.getFiles)
}

func (i *FrameworkController) getVersions(ctx *gin.Context) {
	response := gin.H{
		"latestVersion": "",
	}
	framework := ctx.Param("framework")
	dirs, err := os.ReadDir(i.baseDir + "/" + framework)
	if err == nil {
		dirs = lo.Filter(dirs, func(dir os.DirEntry, _ int) bool {
			return dir.IsDir()
		})
		sort.Slice(dirs, func(i, j int) bool {
			f1, e1 := strconv.ParseFloat(dirs[i].Name(), 32)
			f2, e2 := strconv.ParseFloat(dirs[j].Name(), 32)
			if e1 != nil || e2 != nil {
				return false
			}
			return f1 > f2
		})
		response["latestVersion"] = dirs[0].Name()
		response["versions"] = lo.Map(dirs, func(dir os.DirEntry, _ int) string {
			return dir.Name()
		})
	}
	ctx.JSON(http.StatusOK, response)
}

func (i *FrameworkController) getFileInfo(ctx *gin.Context) {
	framework := ctx.Param("framework")
	version := ctx.Param("version")
	fileDir := i.baseDir + "/" + framework + "/" + version

	fileInfo := []string{}

	if _, err := os.Stat(fileDir); err == nil {
		// Function to transform absolute path to relative path
		transform := func(path string) string {
			return filepath.Base(path)
		}

		// Process all files recursively
		err := filepath.Walk(fileDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, ".DS_Store") {
				fileInfo = append(fileInfo, transform(path))
			}

			return nil
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"files": fileInfo,
	})

}

func (i *FrameworkController) getFiles(ctx *gin.Context) {
	framework := ctx.Param("framework")
	version := ctx.Param("version")
	filePath := ctx.Param("filepath")

	// 打印收到的参数，便于调试
	fmt.Printf("Received params - framework: %s, version: %s, filepath: %s\n", framework, version, filePath)

	if filePath == "" || filePath == "/" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file path is required"})
		return
	}

	// 去掉路径开头的斜杠
	filePath = strings.TrimPrefix(filePath, "/")

	// 检查目录是否存在
	basePath := filepath.Join(i.baseDir, framework)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		// 尝试直接作为文件名处理
		basePath = i.baseDir
	}

	// 构建完整路径，使用filepath.Join确保路径格式正确
	fullPath := filepath.Join(basePath, version, filePath)

	// 打印实际查找的路径，便于调试
	fmt.Printf("Attempting to access file: %s\n", fullPath)

	fileInfo, err := os.Stat(fullPath)

	if err != nil {
		if os.IsNotExist(err) {
			// 提供更具体的错误信息
			ctx.JSON(http.StatusNotFound, gin.H{
				"error":   "file not found",
				"path":    fullPath,
				"details": err.Error(),
			})
		} else {
			// 其他类型的错误
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   "unable to access file",
				"details": err.Error(),
			})
		}
		return
	}

	if fileInfo.IsDir() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot download a directory"})
		return
	}

	fileName := filepath.Base(fullPath)

	// 设置响应头
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)

	// 发送文件
	ctx.File(fullPath)
}
