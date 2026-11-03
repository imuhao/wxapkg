export namespace main {
	
	export class FileFilter {
	    DisplayName: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DisplayName = source["DisplayName"];
	        this.Pattern = source["Pattern"];
	    }
	}

}

export namespace wechat {
	
	export class ExtStat {
	    Ext: string;
	    Count: number;
	    Size: number;
	
	    static createFrom(source: any = {}) {
	        return new ExtStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Ext = source["Ext"];
	        this.Count = source["Count"];
	        this.Size = source["Size"];
	    }
	}
	export class FileStat {
	    Path: string;
	    Size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Size = source["Size"];
	    }
	}
	export class PluginInfo {
	    Alias: string;
	    Provider: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Alias = source["Alias"];
	        this.Provider = source["Provider"];
	        this.Version = source["Version"];
	    }
	}
	export class TabBarItemInfo {
	    Text: string;
	    PagePath: string;
	
	    static createFrom(source: any = {}) {
	        return new TabBarItemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Text = source["Text"];
	        this.PagePath = source["PagePath"];
	    }
	}
	export class SubPackageInfo {
	    Root: string;
	    Name: string;
	    Independent: boolean;
	    PageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new SubPackageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Root = source["Root"];
	        this.Name = source["Name"];
	        this.Independent = source["Independent"];
	        this.PageCount = source["PageCount"];
	    }
	}
	export class AppInfo {
	    ConfigFile: string;
	    AppId: string;
	    AppName: string;
	    AppIcon: string;
	    EntryPage: string;
	    NavigationBarTitle: string;
	    Pages: string[];
	    SubPackages: SubPackageInfo[];
	    TabBar: TabBarItemInfo[];
	    Plugins: PluginInfo[];
	    Permissions: string[];
	    RequiredPrivateInfos: string[];
	    RequiredBackground: string[];
	    Frameworks: string[];
	    FileCount: number;
	    TotalSize: number;
	    ExtStats: ExtStat[];
	    LargeFiles: FileStat[];
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ConfigFile = source["ConfigFile"];
	        this.AppId = source["AppId"];
	        this.AppName = source["AppName"];
	        this.AppIcon = source["AppIcon"];
	        this.EntryPage = source["EntryPage"];
	        this.NavigationBarTitle = source["NavigationBarTitle"];
	        this.Pages = source["Pages"];
	        this.SubPackages = this.convertValues(source["SubPackages"], SubPackageInfo);
	        this.TabBar = this.convertValues(source["TabBar"], TabBarItemInfo);
	        this.Plugins = this.convertValues(source["Plugins"], PluginInfo);
	        this.Permissions = source["Permissions"];
	        this.RequiredPrivateInfos = source["RequiredPrivateInfos"];
	        this.RequiredBackground = source["RequiredBackground"];
	        this.Frameworks = source["Frameworks"];
	        this.FileCount = source["FileCount"];
	        this.TotalSize = source["TotalSize"];
	        this.ExtStats = this.convertValues(source["ExtStats"], ExtStat);
	        this.LargeFiles = this.convertValues(source["LargeFiles"], FileStat);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PathScanResult {
	    Paths: string[];
	    Logs: string;
	
	    static createFrom(source: any = {}) {
	        return new PathScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Paths = source["Paths"];
	        this.Logs = source["Logs"];
	    }
	}
	export class UnpackOptions {
	    EnableDecrypt: boolean;
	    EnableJsBeautify: boolean;
	    EnableHtmlBeautify: boolean;
	    EnableJsonBeautify: boolean;
	    OutputDir: string;
	    SavePath: string;
	
	    static createFrom(source: any = {}) {
	        return new UnpackOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.EnableDecrypt = source["EnableDecrypt"];
	        this.EnableJsBeautify = source["EnableJsBeautify"];
	        this.EnableHtmlBeautify = source["EnableHtmlBeautify"];
	        this.EnableJsonBeautify = source["EnableJsonBeautify"];
	        this.OutputDir = source["OutputDir"];
	        this.SavePath = source["SavePath"];
	    }
	}
	export class WxapkgItem {
	    UUID: string;
	    WxId: string;
	    Location: string;
	    EncryptKey: string;
	    Size: number;
	    IsDir: boolean;
	    LastModifyTime: number;
	    WxapkgFilePaths: string[];
	    UnpackStatus: string;
	    UnpackCurrent: number;
	    UnpackTotal: number;
	    UnpackProgress: number;
	    UnpackCurrentFile: string;
	    UnpackSavePath: string;
	    UnpackErrorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new WxapkgItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UUID = source["UUID"];
	        this.WxId = source["WxId"];
	        this.Location = source["Location"];
	        this.EncryptKey = source["EncryptKey"];
	        this.Size = source["Size"];
	        this.IsDir = source["IsDir"];
	        this.LastModifyTime = source["LastModifyTime"];
	        this.WxapkgFilePaths = source["WxapkgFilePaths"];
	        this.UnpackStatus = source["UnpackStatus"];
	        this.UnpackCurrent = source["UnpackCurrent"];
	        this.UnpackTotal = source["UnpackTotal"];
	        this.UnpackProgress = source["UnpackProgress"];
	        this.UnpackCurrentFile = source["UnpackCurrentFile"];
	        this.UnpackSavePath = source["UnpackSavePath"];
	        this.UnpackErrorMessage = source["UnpackErrorMessage"];
	    }
	}

}

