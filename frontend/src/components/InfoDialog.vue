<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import * as AppService from '../../wailsjs/go/main/AppService'
import { wechat } from '../../wailsjs/go/models'
import { formatSize } from '../entries/util'
import AppInfo = wechat.AppInfo

const props = defineProps<{
  unpackedDir: string;
  wxId: string;
}>()

const visible = defineModel<boolean>('visible', { default: false })
const info = ref<AppInfo | null>(null)
const loading = ref(false)
const error = ref('')
const showAllPages = ref(false)

watch(visible, (v) => {
  if (!v || !props.unpackedDir) return
  info.value = null
  error.value = ''
  showAllPages.value = false
  loading.value = true
  AppService.GetAppInfo(props.unpackedDir)
    .then(res => info.value = res)
    .catch(e => error.value = String(e))
    .finally(() => loading.value = false)
})

const totalPages = computed(() => {
  if (!info.value) return 0
  const sub = (info.value.SubPackages ?? []).reduce((n, s) => n + s.PageCount, 0)
  return (info.value.Pages?.length ?? 0) + sub
})

const visiblePages = computed(() => {
  const pages = info.value?.Pages ?? []
  return showAllPages.value ? pages : pages.slice(0, 12)
})

const permissionNames: Record<string, string> = {
  'scope.userLocation': '地理位置',
  'scope.userFuzzyLocation': '模糊地理位置',
  'scope.userLocationBackground': '后台定位',
  'scope.record': '麦克风',
  'scope.camera': '摄像头',
  'scope.bluetooth': '蓝牙',
  'scope.writePhotosAlbum': '保存到相册',
  'scope.addPhoneContact': '添加通讯录',
  'scope.addPhoneCalendar': '添加日历',
  'scope.werun': '微信运动步数',
}

function permLabel(scope: string) {
  return permissionNames[scope] ? `${permissionNames[scope]}（${scope}）` : scope
}

function copy(text: string) {
  AppService.ClipboardSetText(text)
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    header="小程序信息"
    :style="{ width: '640px' }"
    :autofocus="false"
  >
    <div v-if="loading" class="info-state"><i class="pi pi-spin pi-spinner"></i> 正在分析…</div>
    <div v-else-if="error" class="info-state info-error">{{ error }}</div>

    <div v-else-if="info" class="info-body">
      <!-- 基本信息 -->
      <div class="info-section">
        <div class="section-label">基本信息</div>
        <div class="kv">
          <span>小程序 ID</span><span class="mono" @click="copy(info.AppId || wxId)">{{ info.AppId || wxId || '—' }}</span>
          <span v-if="info.AppName">名称</span><span v-if="info.AppName">{{ info.AppName }}</span>
          <span v-if="info.NavigationBarTitle">默认标题</span><span v-if="info.NavigationBarTitle">{{ info.NavigationBarTitle }}</span>
          <span>入口页面</span><span class="mono">{{ info.EntryPage || '—' }}</span>
          <span>配置文件</span><span class="mono">{{ info.ConfigFile || '未找到' }}</span>
        </div>
      </div>

      <!-- 统计 -->
      <div class="stats">
        <div class="stat"><b>{{ totalPages }}</b><span>页面</span></div>
        <div class="stat"><b>{{ info.SubPackages?.length ?? 0 }}</b><span>分包</span></div>
        <div class="stat"><b>{{ info.Plugins?.length ?? 0 }}</b><span>插件</span></div>
        <div class="stat"><b>{{ info.FileCount }}</b><span>文件</span></div>
        <div class="stat"><b>{{ formatSize(info.TotalSize) }}</b><span>总大小</span></div>
      </div>

      <!-- 技术栈 -->
      <div class="info-section">
        <div class="section-label">技术栈（特征识别）</div>
        <div v-if="info.Frameworks?.length" class="tags">
          <span v-for="f in info.Frameworks" :key="f" class="tag">{{ f }}</span>
        </div>
        <div v-else class="muted">未识别到常见框架，可能为原生开发</div>
      </div>

      <!-- TabBar -->
      <div v-if="info.TabBar?.length" class="info-section">
        <div class="section-label">TabBar</div>
        <div class="list">
          <div v-for="t in info.TabBar" :key="t.PagePath" class="row">
            <span>{{ t.Text || '（无文字）' }}</span><span class="mono muted">{{ t.PagePath }}</span>
          </div>
        </div>
      </div>

      <!-- 主包页面 -->
      <div v-if="info.Pages?.length" class="info-section">
        <div class="section-label">主包页面（{{ info.Pages.length }}）</div>
        <div class="list">
          <div v-for="p in visiblePages" :key="p" class="row mono">{{ p }}</div>
        </div>
        <button v-if="info.Pages.length > 12" class="btn-text" @click="showAllPages = !showAllPages">
          {{ showAllPages ? '收起' : `显示全部 ${info.Pages.length} 个` }}
        </button>
      </div>

      <!-- 分包 -->
      <div v-if="info.SubPackages?.length" class="info-section">
        <div class="section-label">分包</div>
        <div class="list">
          <div v-for="s in info.SubPackages" :key="s.Root" class="row">
            <span class="mono">{{ s.Root }}</span>
            <span class="muted">{{ s.PageCount }} 页{{ s.Independent ? ' · 独立分包' : '' }}</span>
          </div>
        </div>
      </div>

      <!-- 插件 -->
      <div v-if="info.Plugins?.length" class="info-section">
        <div class="section-label">插件</div>
        <div class="list">
          <div v-for="p in info.Plugins" :key="p.Alias" class="row">
            <span>{{ p.Alias }}</span><span class="mono muted">{{ p.Provider }} {{ p.Version }}</span>
          </div>
        </div>
      </div>

      <!-- 权限 -->
      <div v-if="info.Permissions?.length || info.RequiredPrivateInfos?.length || info.RequiredBackground?.length" class="info-section">
        <div class="section-label">声明的权限与接口</div>
        <div class="tags">
          <span v-for="p in info.Permissions" :key="p" class="tag">{{ permLabel(p) }}</span>
          <span v-for="p in info.RequiredPrivateInfos" :key="p" class="tag">{{ p }}</span>
          <span v-for="p in info.RequiredBackground" :key="p" class="tag">后台：{{ p }}</span>
        </div>
      </div>

      <!-- 文件类型 -->
      <div class="info-section">
        <div class="section-label">文件类型</div>
        <div class="list">
          <div v-for="e in info.ExtStats" :key="e.Ext" class="row">
            <span class="mono">{{ e.Ext }}</span><span class="muted">{{ e.Count }} 个 · {{ formatSize(e.Size) }}</span>
          </div>
        </div>
      </div>

      <!-- 最大文件 -->
      <div v-if="info.LargeFiles?.length" class="info-section">
        <div class="section-label">最大的文件</div>
        <div class="list">
          <div v-for="f in info.LargeFiles" :key="f.Path" class="row">
            <span class="mono ellipsis">{{ f.Path }}</span><span class="muted">{{ formatSize(f.Size) }}</span>
          </div>
        </div>
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
.info-state { padding: 32px 0; text-align: center; color: var(--color-text-tertiary); }
.info-error { color: #d93025; white-space: pre-wrap; }
.info-body { max-height: 62vh; overflow-y: auto; padding-right: 4px; }
.info-section { margin-bottom: 18px; }
.kv { display: grid; grid-template-columns: 90px 1fr; gap: 6px 12px; font-size: 13px; }
.kv > span:nth-child(odd) { color: var(--color-text-tertiary); }
.kv .mono { cursor: copy; }
.stats { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; margin-bottom: 18px; }
.stat { background: rgba(0,0,0,0.03); border-radius: 8px; padding: 10px; text-align: center; }
.stat b { display: block; font-size: 16px; }
.stat span { font-size: 11px; color: var(--color-text-tertiary); }
.tags { display: flex; flex-wrap: wrap; gap: 6px; }
.tag { font-size: 12px; padding: 3px 8px; border-radius: 6px; background: rgba(0,0,0,0.05); }
.list { border: 1px solid rgba(0,0,0,0.06); border-radius: 8px; }
.row { display: flex; justify-content: space-between; gap: 12px; padding: 6px 10px; font-size: 13px; border-bottom: 1px solid rgba(0,0,0,0.04); }
.row:last-child { border-bottom: none; }
.muted { color: var(--color-text-tertiary); }
.mono { font-family: "JetBrains Mono", "Cascadia Code", "Consolas", monospace; font-size: 12px; }
.ellipsis { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
</style>
