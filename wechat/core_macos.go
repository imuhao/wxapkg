//go:build darwin

package wechat

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 匹配 com.tencent.xinWeChat 以及多开副本 com.tencent.xinWeChat2、com.tencent.xinWeChat3 ...
var reWeChatContainer = regexp.MustCompile(`^com\.tencent\.xinWeChat\d*$`)

type macosPlatform struct{}

func newPlatform() platform { return &macosPlatform{} }

func (m *macosPlatform) GetDefaultPaths() PathScanResult {
	var b strings.Builder
	log := func(line string) { b.WriteString(line + "\n") }

	var paths []string

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log("获取用户目录失败: " + err.Error())
		return PathScanResult{Paths: paths, Logs: b.String()}
	}

	containersDir := filepath.Join(userHomeDir, "Library/Containers")
	containerEntries, err := os.ReadDir(containersDir)
	if err != nil {
		log(fmt.Sprintf("读取目录失败: %s, %v", containersDir, err))
		return PathScanResult{Paths: paths, Logs: b.String()}
	}

	var containers []string
	for _, entry := range containerEntries {
		if entry.IsDir() && reWeChatContainer.MatchString(entry.Name()) {
			containers = append(containers, entry.Name())
		}
	}
	if len(containers) == 0 {
		log(fmt.Sprintf("未找到微信容器目录 (com.tencent.xinWeChat*)\n   路径: %s", containersDir))
		return PathScanResult{Paths: paths, Logs: b.String()}
	}

	checkPath := func(step, desc, p string) {
		if fileInfo, err := os.Stat(p); err != nil || !fileInfo.IsDir() {
			log(fmt.Sprintf("%s 【失败】检测%s - 目录不存在\n   路径: %s", step, desc, p))
		} else if !hasWxapkgApp(p) {
			log(fmt.Sprintf("%s 【失败】检测%s - 目录中没有小程序\n   路径: %s", step, desc, p))
		} else {
			log(fmt.Sprintf("%s 【成功】检测%s - 成功\n   路径: %s", step, desc, p))
			paths = append(paths, p)
		}
	}

	for i, container := range containers {
		containerData := filepath.Join(containersDir, container, "Data")
		log(fmt.Sprintf("[%s]", container))

		// ── Step 1: 微信 v3 ──
		checkPath(fmt.Sprintf("%d.1", i+1), "微信 v3 (3.x 版本)", filepath.Join(containerData, ".wxapplet/packages"))

		// ── Step 2: 微信 v4 ──
		checkPath(fmt.Sprintf("%d.2", i+1), "微信 v4 (4.x 版本)", filepath.Join(containerData, "Documents/app_data/radium/Applet/packages"))

		// ── Step 3: 微信 v4 多用户 ──
		v4UsersPath := filepath.Join(containerData, "Documents/app_data/radium/users")
		entries, err := os.ReadDir(v4UsersPath)
		if err != nil {
			log(fmt.Sprintf("%d.3 【失败】检测微信 v4 多用户 (4.x 版本) - 目录不存在\n   路径: %s", i+1, v4UsersPath))
			continue
		}
		var found []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			userAppletPath := filepath.Join(v4UsersPath, entry.Name(), "applet", "packages")
			if hasWxapkgApp(userAppletPath) {
				paths = append(paths, userAppletPath)
				found = append(found, entry.Name())
			}
		}
		if len(found) == 0 {
			log(fmt.Sprintf("%d.3 【失败】检测微信 v4 多用户 (4.x 版本) - 未找到包含小程序的用户\n   路径: %s", i+1, v4UsersPath))
		} else {
			log(fmt.Sprintf("%d.3 【成功】检测微信 v4 多用户 (4.x 版本) - 成功\n   路径: %s\n   有效用户: %s", i+1, v4UsersPath, strings.Join(found, ", ")))
		}
	}

	return PathScanResult{Paths: paths, Logs: b.String()}
}
