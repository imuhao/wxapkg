<script setup lang="ts">
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import * as AppService from '../../wailsjs/go/main/AppService'

const props = defineProps<{
  defaultOutputDir: string;
}>()

const emit = defineEmits<{
  save: [path: string];
}>()

const visible = defineModel<boolean>('visible', { default: false })
const draftOutputDir = ref('')

watch(visible, (newVal) => {
  if (newVal) {
    draftOutputDir.value = props.defaultOutputDir ?? ''
  }
})

function selectDirectory() {
  AppService.OpenDirectoryDialog('选择默认输出目录', draftOutputDir.value)
    .then((path) => {
      if (path) {
        draftOutputDir.value = path
      }
    })
}

function clearDirectory() {
  draftOutputDir.value = ''
}

function saveSettings() {
  emit('save', draftOutputDir.value.trim())
  visible.value = false
}
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    header="设置"
    :style="{ width: '520px' }"
    :autofocus="false"
  >
    <div class="form-section">
      <div class="section-label">默认输出目录</div>
      <div class="form-input-wrapper">
        <input
          v-model="draftOutputDir"
          class="form-input"
          type="text"
          placeholder="可手动输入，或点击右侧图标选择目录"
          style="width:100%; padding-right:40px"
        />
        <i
          class="pi pi-folder form-input-icon-right"
          style="pointer-events:auto; cursor:pointer"
          @click="selectDirectory"
        ></i>
      </div>
      <p class="form-hint">
        未设置时，解包弹窗会优先回填上一次输出目录。
      </p>
    </div>

    <div class="dialog-footer" style="padding: 16px 0 0; border: none; justify-content:flex-end; gap:8px">
      <button class="btn-secondary" @click="clearDirectory">清空默认目录</button>
      <button class="btn-primary" @click="saveSettings">保存</button>
    </div>
  </Dialog>
</template>

<style scoped>
.form-section {
  margin-bottom: 4px;
}

.form-hint {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 8px;
}
</style>
