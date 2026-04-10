<template>
  <div class="markdown-view" :class="{ 'is-dark': resolvedTheme === 'dark' }">
    <MdPreview
      :modelValue="renderedSource"
      :editorId="editorId"
      :theme="viewerTheme"
      :previewTheme="viewerPreviewTheme"
      :codeTheme="viewerCodeTheme"
      languageUserDefined
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { MdPreview } from 'md-editor-v3';
import 'md-editor-v3/lib/preview.css';
import { useTheme } from '@/composables/useTheme';

const props = defineProps<{
  source?: string;
}>();
const { resolvedTheme } = useTheme();
const editorId = 'analysis-preview';
const viewerTheme = resolvedTheme.value === 'dark' ? 'dark' : 'light';
const viewerPreviewTheme = resolvedTheme.value === 'dark' ? 'default' : 'github';
const viewerCodeTheme = 'atom';

const renderedSource = computed(() => {
  const input = normalizeSource(props.source || '');
  if (!input) {
    return '暂无解析';
  }

  return input
    .replace(/\\\[([\s\S]*?)\\\]/g, (_m, p1) => `$$${String(p1).trim()}$$`)
    .replace(/\\\(([^\n]*?)\\\)/g, (_m, p1) => `$${String(p1).trim()}$`);
});

function normalizeSource(raw: string) {
  let input = decodeHtml(String(raw || '').trim());
  if (!input) {
    return '';
  }

  input = cleanupText(input)
    .replace(/,\s*,+/g, ', ')
    .replace(/，\s*，+/g, '，')
    .replace(/(\(|\[)\s*,\s*/g, '$1')
    .replace(/\s*,\s*(\)|\])/g, '$1');

  input = input.replace(/(^|\n)(\\begin\{cases\}[\s\S]*?\\end\{cases\})(?=\n|$)/g, (_m, p1, p2) => {
    const block = String(p2 || '').trim();
    return `${p1}\n$$${block}$$\n`;
  });

  return input;
}

function decodeHtml(input: string) {
  return input
    .replace(/&nbsp;/gi, ' ')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&amp;/gi, '&')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/gi, "'")
    .replace(/&ZeroWidthSpace;/gi, '');
}

function cleanupText(input: string) {
  return input
    .replace(/\r/g, '')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\n[ \t]+/g, '\n')
    .replace(/[ \t]{2,}/g, ' ')
    .replace(/\n{3,}/g, '\n\n')
    .trim();
}
</script>

<style scoped>
.markdown-view {
  line-height: 1.72;
  font-size: 14px;
  color: #334155;
  overflow-x: auto;
}

.markdown-view :deep(.md-editor-preview-wrapper) {
  padding: 0;
}

.markdown-view :deep(.md-editor-preview) {
  background: transparent;
  padding: 0;
}

.markdown-view :deep(.md-editor-preview p) {
  margin: 0;
}

.markdown-view :deep(.md-editor-preview p + p) {
  margin-top: 8px;
}

.markdown-view :deep(.md-editor-preview ul),
.markdown-view :deep(.md-editor-preview ol) {
  margin: 8px 0 0;
  padding-left: 20px;
}

.markdown-view :deep(.md-editor-preview blockquote) {
  margin: 8px 0 0;
  border-left: 3px solid rgba(59, 130, 246, 0.42);
  padding-left: 10px;
  color: #475569;
}

.markdown-view :deep(code) {
  background: rgba(148, 163, 184, 0.16);
  border-radius: 6px;
  padding: 1px 5px;
}

.markdown-view :deep(.katex) {
  color: inherit;
}

.markdown-view :deep(.katex-display) {
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 100%;
  margin: 0.6em 0;
}

.markdown-view :deep(.katex-display > .katex) {
  white-space: nowrap;
}

.markdown-view :deep(.katex-render) {
  display: inline-block;
}

.markdown-view.is-dark {
  color: #cbd5e1;
}

.markdown-view.is-dark :deep(.md-editor) {
  background: transparent;
}

.markdown-view.is-dark :deep(.md-editor-preview-wrapper) {
  background: transparent;
}

.markdown-view.is-dark :deep(.katex) {
  color: #cbd5e1;
}
</style>
