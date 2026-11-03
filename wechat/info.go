package wechat

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// AppInfo 解包后小程序的结构性概览信息
type AppInfo struct {
	ConfigFile string // 读取到的配置文件（相对解包目录）

	AppId     string
	AppName   string
	AppIcon   string
	EntryPage string

	NavigationBarTitle string

	Pages       []string
	SubPackages []SubPackageInfo
	TabBar      []TabBarItemInfo
	Plugins     []PluginInfo

	Permissions          []string // permission 中声明的权限 scope
	RequiredPrivateInfos []string
	RequiredBackground   []string

	Frameworks []string // 识别出的开发框架 / 组件库

	FileCount  int
	TotalSize  int64
	ExtStats   []ExtStat
	LargeFiles []FileStat
}

type SubPackageInfo struct {
	Root        string
	Name        string
	Independent bool
	PageCount   int
}

type TabBarItemInfo struct {
	Text     string
	PagePath string
}

type PluginInfo struct {
	Alias    string
	Provider string
	Version  string
}

type ExtStat struct {
	Ext   string
	Count int
	Size  int64
}

type FileStat struct {
	Path string
	Size int64
}

// 框架 / 组件库特征，在 js/wxss/json 文件中匹配
var frameworkSignatures = []struct {
	name     string
	patterns []string
}{
	{"uni-app", []string{"__uniConfig", "uni-app", "__uniRoutes"}},
	{"Taro", []string{"@tarojs", "taro-runtime", "__taroPluginMap"}},
	{"mpvue", []string{"mpvue"}},
	{"WePY", []string{"wepy"}},
	{"kbone", []string{"miniprogram-render", "kbone"}},
	{"Remax", []string{"@remax/"}},
	{"Chameleon", []string{"chameleon-runtime"}},
	{"Vant Weapp", []string{"van-button", "@vant/weapp", "vant-weapp"}},
	{"TDesign", []string{"tdesign-miniprogram", "t-button"}},
	{"iView Weapp", []string{"i-button", "iview-weapp"}},
	{"ColorUI", []string{"colorui", "cu-btn"}},
	{"MobX", []string{"mobx-miniprogram"}},
	{"Lodash", []string{"lodash"}},
	{"Moment.js", []string{"moment.js", "momentProperties"}},
	{"ECharts", []string{"echarts"}},
}

const maxSignatureScanSize = 4 * 1024 * 1024

// AnalyzeUnpacked 解析解包目录，提取小程序的结构性概览信息
func AnalyzeUnpacked(dir string) (*AppInfo, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, errors.Errorf("解包目录 '%s' 不存在", dir)
	}

	info := &AppInfo{}
	if err := info.collectFiles(dir); err != nil {
		return nil, errors.Errorf("遍历解包目录失败，%v", err)
	}

	cfgPath := findAppConfig(dir)
	if cfgPath != "" {
		rel, _ := filepath.Rel(dir, cfgPath)
		info.ConfigFile = filepath.ToSlash(rel)
		if err := info.parseConfig(cfgPath); err != nil {
			return nil, errors.Errorf("解析配置文件 '%s' 失败，%v", info.ConfigFile, err)
		}
	}

	return info, nil
}

func (info *AppInfo) collectFiles(dir string) error {
	extMap := map[string]*ExtStat{}
	var files []FileStat
	found := map[string]bool{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		rel = filepath.ToSlash(rel)

		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = "(无)"
		}
		s, ok := extMap[ext]
		if !ok {
			s = &ExtStat{Ext: ext}
			extMap[ext] = s
		}
		s.Count++
		s.Size += fi.Size()

		info.FileCount++
		info.TotalSize += fi.Size()
		files = append(files, FileStat{Path: rel, Size: fi.Size()})

		if (ext == ".js" || ext == ".json" || ext == ".wxss" || ext == ".wxml" || ext == ".html") && fi.Size() <= maxSignatureScanSize {
			detectFrameworks(path, found)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, s := range extMap {
		info.ExtStats = append(info.ExtStats, *s)
	}
	sort.Slice(info.ExtStats, func(i, j int) bool { return info.ExtStats[i].Size > info.ExtStats[j].Size })

	sort.Slice(files, func(i, j int) bool { return files[i].Size > files[j].Size })
	if len(files) > 10 {
		files = files[:10]
	}
	info.LargeFiles = files

	for _, sig := range frameworkSignatures {
		if found[sig.name] {
			info.Frameworks = append(info.Frameworks, sig.name)
		}
	}
	return nil
}

func detectFrameworks(path string, found map[string]bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, sig := range frameworkSignatures {
		if found[sig.name] {
			continue
		}
		for _, p := range sig.patterns {
			if bytes.Contains(data, []byte(p)) {
				found[sig.name] = true
				break
			}
		}
	}
}

// findAppConfig 优先查找根目录的 app-config.json（编译后的配置），其次 app.json
func findAppConfig(dir string) string {
	for _, name := range []string{"app-config.json", "app.json"} {
		p := filepath.Join(dir, name)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	// 退而求其次：在浅层目录中查找
	var result string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if result != "" {
			return filepath.SkipAll
		}
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if d.IsDir() && strings.Count(filepath.ToSlash(rel), "/") >= 2 {
			return filepath.SkipDir
		}
		if !d.IsDir() && (d.Name() == "app-config.json" || d.Name() == "app.json") {
			result = path
		}
		return nil
	})
	return result
}

type rawAppConfig struct {
	EntryPagePath string   `json:"entryPagePath"`
	Pages         []string `json:"pages"`
	SubPackages   []rawSub `json:"subPackages"`
	SubPackages2  []rawSub `json:"subpackages"`
	Global        struct {
		Window map[string]any `json:"window"`
	} `json:"global"`
	Window map[string]any `json:"window"`
	TabBar struct {
		List []struct {
			Text     string `json:"text"`
			PagePath string `json:"pagePath"`
		} `json:"list"`
	} `json:"tabBar"`
	Plugins              map[string]rawPlugin `json:"plugins"`
	Permission           map[string]any       `json:"permission"`
	RequiredPrivateInfos []string             `json:"requiredPrivateInfos"`
	RequiredBackground   []string             `json:"requiredBackgroundModes"`
	AccountInfo          struct {
		AppId    string `json:"appId"`
		Nickname string `json:"nickname"`
		Icon     string `json:"icon"`
	} `json:"accountInfo"`
	AppId string `json:"appid"`
}

type rawSub struct {
	Root        string   `json:"root"`
	Name        string   `json:"name"`
	Pages       []string `json:"pages"`
	Independent bool     `json:"independent"`
}

type rawPlugin struct {
	Provider string `json:"provider"`
	Version  string `json:"version"`
}

func (info *AppInfo) parseConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg rawAppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	info.AppId = cfg.AccountInfo.AppId
	if info.AppId == "" {
		info.AppId = cfg.AppId
	}
	info.AppName = cfg.AccountInfo.Nickname
	info.AppIcon = cfg.AccountInfo.Icon
	info.EntryPage = trimPageSuffix(cfg.EntryPagePath)

	window := cfg.Global.Window
	if window == nil {
		window = cfg.Window
	}
	if title, ok := window["navigationBarTitleText"].(string); ok {
		info.NavigationBarTitle = title
	}

	for _, p := range cfg.Pages {
		info.Pages = append(info.Pages, trimPageSuffix(p))
	}
	if info.EntryPage == "" && len(info.Pages) > 0 {
		info.EntryPage = info.Pages[0]
	}

	subs := cfg.SubPackages
	if len(subs) == 0 {
		subs = cfg.SubPackages2
	}
	for _, s := range subs {
		info.SubPackages = append(info.SubPackages, SubPackageInfo{
			Root:        s.Root,
			Name:        s.Name,
			Independent: s.Independent,
			PageCount:   len(s.Pages),
		})
	}

	for _, t := range cfg.TabBar.List {
		info.TabBar = append(info.TabBar, TabBarItemInfo{Text: t.Text, PagePath: trimPageSuffix(t.PagePath)})
	}

	for alias, p := range cfg.Plugins {
		info.Plugins = append(info.Plugins, PluginInfo{Alias: alias, Provider: p.Provider, Version: p.Version})
	}
	sort.Slice(info.Plugins, func(i, j int) bool { return info.Plugins[i].Alias < info.Plugins[j].Alias })

	for scope := range cfg.Permission {
		info.Permissions = append(info.Permissions, scope)
	}
	sort.Strings(info.Permissions)

	info.RequiredPrivateInfos = cfg.RequiredPrivateInfos
	info.RequiredBackground = cfg.RequiredBackground
	return nil
}

func trimPageSuffix(p string) string {
	return strings.TrimSuffix(p, ".html")
}
