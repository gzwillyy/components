package fileutil

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

// FileType 使用文件类型包来确定给定文件路径的类型.
func FileType(filePath string) (types.Type, error) {
	file, _ := os.Open(filePath)

	// We only have to pass the file header = first 261 bytes
	head := make([]byte, 261)
	_, _ = file.Read(head)

	return filetype.Match(head)
}

// FileExists 如果给定路径存在则返回 true.
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	return false, err
}

// DirExists 如果给定路径存在并且是目录，则返回 true.
func DirExists(path string) (bool, error) {
	exists, _ := FileExists(path)
	fileInfo, _ := os.Stat(path)
	if !exists || !fileInfo.IsDir() {
		return false, fmt.Errorf("path either doesn't exist, or is not a directory <%s>", path)
	}
	return true, nil
}

// Touch 如果给定路径不存在则创建一个空文件.
func Touch(path string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

// EnsureDir 如果目录不存在，将在给定路径创建一个目录.
func EnsureDir(path string) error {
	exists, err := FileExists(path)
	if !exists {
		err = os.Mkdir(path, 0755)
		return err
	}
	return err
}

// EnsureDirAll 将在给定路径创建目录以及任何必要的父目录（如果它们尚不存在）.
func EnsureDirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// RemoveDir 删除给定的目录（如果存在）及其所有内容.
func RemoveDir(path string) error {
	return os.RemoveAll(path)
}

// EmptyDir 将递归地删除给定路径下目录的内容.
func EmptyDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}

	return nil
}

// ListDir 将以字符串切片的形式返回给定目录路径的内容.
func ListDir(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		path = filepath.Dir(path)
		files, _ = ioutil.ReadDir(path)
	}

	//nolint: prealloc
	var dirPaths []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		dirPaths = append(dirPaths, filepath.Join(path, file.Name()))
	}
	return dirPaths
}

// GetHomeDirectory 返回用户主目录的路径. ~ 在 Unix 上和 C:\Users\UserName 在 Windows 上.
func GetHomeDirectory() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return currentUser.HomeDir
}

// SafeMove 在安全模式下将 src 移动到 dst.
func SafeMove(src, dst string) error {
	err := os.Rename(src, dst)

	//nolint: nestif
	if err != nil {
		fmt.Printf("[fileutil] unable to rename: \"%s\" due to %s. Falling back to copying.", src, err.Error())

		in, err := os.Open(src)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}

		err = out.Close()
		if err != nil {
			return err
		}

		err = os.Remove(src)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsZipFileUncompressed 如果路径中的 zip 文件使用 0 压缩级别，则返回 true.
func IsZipFileUncompressed(path string) (bool, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		fmt.Printf("Error reading zip file %s: %s\n", path, err)
		return false, err
	}
	defer r.Close()
	for _, f := range r.File {
		if f.FileInfo().IsDir() { // skip dirs, they always get store level compression
			continue
		}
		return f.Method == 0, nil // check compression level of first actual  file
	}
	return false, nil
}

// WriteFile 如果需要，将文件写入创建父目录的路径.
func WriteFile(path string, file []byte) error {
	pathErr := EnsureDirAll(filepath.Dir(path))
	if pathErr != nil {
		return fmt.Errorf("cannot ensure path %s", pathErr)
	}

	err := ioutil.WriteFile(path, file, 0600)
	if err != nil {
		return fmt.Errorf("write error for thumbnail %s: %s ", path, err)
	}
	return nil
}

// GetIntraDir 返回一个字符串，可以添加到 filepath.Join 以实现目录深度，
// “”在错误时例如给定模式 0af63ce3c99162e9df23a997f62621c5 和深度 2 长度 3
// 返回 0af/63c 或 0af\63c（取决于操作系统） 以后可以像这样使用 filepath.Join(directory, intradir, basename).
func GetIntraDir(pattern string, depth, length int) string {
	if depth < 1 || length < 1 || (depth*length > len(pattern)) {
		return ""
	}
	intraDir := pattern[0:length] // depth 1 , get length number of characters from pattern
	for i := 1; i < depth; i++ {  // for every extra depth: move to the right of the pattern length positions, get length number of chars
		intraDir = filepath.Join(
			intraDir,
			pattern[length*i:length*(i+1)],
		) //  adding each time to intradir the extra characters with a filepath join
	}
	return intraDir
}

// GetParent 返回给定路径的父目录.
func GetParent(path string) *string {
	isRoot := path[len(path)-1:] == "/"
	if isRoot {
		return nil
	}
	parentPath := filepath.Clean(path + "/..")
	return &parentPath
}

// ServeFileNoCache 提供提供的文件，确保响应包含标题以防止缓存.
func ServeFileNoCache(w http.ResponseWriter, r *http.Request, filepath string) {
	w.Header().Add("Cache-Control", "no-cache")

	http.ServeFile(w, r, filepath)
}

// MatchEntries 返回目录 dir 中与正则表达式模式匹配的条目的字符串切片. 出错时返回一个空切片
// MatchEntries 不是递归的，只搜索特定的“dir”而不展开.
func MatchEntries(dir, pattern string) ([]string, error) {
	var res []string
	var err error

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	files, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if re.Match([]byte(file)) {
			res = append(res, filepath.Join(dir, file))
		}
	}
	return res, err
}
