package wechat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	uuidlib "github.com/google/uuid"
	"golang.org/x/crypto/pbkdf2"
)

func getPathSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func uuid() string {
	newUUID, _ := uuidlib.NewRandom()
	return newUUID.String()
}

// ListFilesWithExtension extension eg: .wxapkg
func ListFilesWithExtension(dir, extension string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, extension) {
			paths = append(paths, path)
		}
		return nil
	})

	return paths, err
}

func decryptWxapkgFile(wxid string, dataByte []byte) ([]byte, error) {
	if len(dataByte) < 1024+6 {
		return nil, fmt.Errorf("文件长度 %d 字节过小，不是有效的加密 wxapkg 文件", len(dataByte))
	}
	var (
		salt = "saltiest"
		iv   = "the iv: 16 bytes"
	)

	dk := pbkdf2.Key([]byte(wxid), []byte(salt), 1000, 32, sha1.New)
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	originData := make([]byte, 1024)
	blockMode.CryptBlocks(originData, dataByte[6:1024+6])

	afData := make([]byte, len(dataByte)-1024-6) // remove first 6 + 1024 byte
	var xorKey = byte(0x66)
	if len(wxid) >= 2 {
		xorKey = wxid[len(wxid)-2]
	}
	for i, b := range dataByte[1024+6:] { // from 6 + 1024 byte
		afData[i] = b ^ xorKey
	}

	originData = append(originData[:1023], afData...)

	return originData, nil
}

func ScanWxapkgItem(path string, scan bool) ([]WxapkgItem, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() || !scan || reWxId.MatchString(filepath.Base(path)) {
		// 选中的是单个小程序目录(wx + 16位)时，自动使用目录名作为 wxid 和解密密钥
		wxId, key := "unknown", ""
		if stat.IsDir() && reWxId.MatchString(filepath.Base(path)) {
			wxId, key = filepath.Base(path), filepath.Base(path)
		}
		return []WxapkgItem{{
			UUID:           uuid(),
			WxId:           wxId,
			EncryptKey:     key,
			Location:       path,
			Size:           getPathSize(path),
			IsDir:          stat.IsDir(),
			LastModifyTime: stat.ModTime().Unix(),
		}}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result []WxapkgItem

	for _, entry := range entries {
		if entry.IsDir() {
			dirName := entry.Name()
			if reWxId.MatchString(dirName) {
				localPath := filepath.Join(path, dirName)

				item := WxapkgItem{
					UUID:       uuid(),
					WxId:       dirName,
					Location:   localPath,
					EncryptKey: dirName,
					Size:       getPathSize(localPath),
					IsDir:      true,
				}
				if info, err := os.Stat(localPath); err == nil {
					item.LastModifyTime = info.ModTime().Unix()
				}

				result = append(result, item)
			}
		}
	}

	return result, nil
}

// hasWxapkgApp 判断目录下是否存在至少一个小程序目录(wx + 16位)
func hasWxapkgApp(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && reWxId.MatchString(entry.Name()) {
			return true
		}
	}
	return false
}

// latestVersionFiles 小程序目录下按版本号分子目录存放(如 wxid/104、wxid/112)，只保留最新版本的文件，
// 避免多个版本合并解包产生大量重复文件。目录结构不符合时原样返回。
func latestVersionFiles(root string, files []string) []string {
	var latest = -1
	var versions = make([]int, len(files))
	for i, file := range files {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return files
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) < 2 {
			return files
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return files
		}
		versions[i] = version
		latest = max(latest, version)
	}

	var result []string
	for i, file := range files {
		if versions[i] == latest {
			result = append(result, file)
		}
	}
	return result
}
