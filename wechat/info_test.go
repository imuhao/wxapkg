package wechat

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeUnpacked(t *testing.T) {
	dir := t.TempDir()
	cfg := `{"entryPagePath":"pages/index/index.html","pages":["pages/index/index.html","pages/me/me.html"],
	"global":{"window":{"navigationBarTitleText":"Demo"}},
	"tabBar":{"list":[{"text":"首页","pagePath":"pages/index/index.html"}]},
	"subPackages":[{"root":"sub/","pages":["a.html","b.html"],"independent":true}],
	"plugins":{"live":{"provider":"wx2b03c6e691cd7370","version":"1.0.0"}},
	"permission":{"scope.userLocation":{"desc":"x"}},
	"requiredPrivateInfos":["getLocation"],
	"accountInfo":{"appId":"wx0123456789abcdef","nickname":"演示"}}`
	os.WriteFile(filepath.Join(dir, "app-config.json"), []byte(cfg), 0600)
	os.MkdirAll(filepath.Join(dir, "common"), 0700)
	os.WriteFile(filepath.Join(dir, "common", "vendor.js"), []byte("var __uniConfig={};"), 0600)

	info, err := AnalyzeUnpacked(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.AppId != "wx0123456789abcdef" || info.AppName != "演示" || info.EntryPage != "pages/index/index" {
		t.Fatalf("basic info: %+v", info)
	}
	if len(info.Pages) != 2 || len(info.SubPackages) != 1 || info.SubPackages[0].PageCount != 2 || !info.SubPackages[0].Independent {
		t.Fatalf("pages/subpackages: %+v", info)
	}
	if len(info.TabBar) != 1 || len(info.Plugins) != 1 || len(info.Permissions) != 1 || len(info.RequiredPrivateInfos) != 1 {
		t.Fatalf("tabbar/plugins/perm: %+v", info)
	}
	if info.NavigationBarTitle != "Demo" || len(info.Frameworks) != 1 || info.Frameworks[0] != "uni-app" || info.FileCount != 2 {
		t.Fatalf("misc: %+v", info)
	}
}
