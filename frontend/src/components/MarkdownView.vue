<template>
  <div class="markdown-view" :class="{ 'is-dark': resolvedTheme === 'dark' }" v-html="html" />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import MarkdownIt from 'markdown-it';
import katex from 'markdown-it-katex';
import 'katex/dist/katex.min.css';
import { useTheme } from '@/composables/useTheme';

const props = defineProps<{
  source?: string;
}>();
const { resolvedTheme } = useTheme();

const md = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
}).use(katex);

const html = computed(() => {
  const input = (props.source || '').trim();
  if (!input) {
    return '<p>暂无解析</p>';
  }
  const normalized = input
    .replace(/\\\[([\s\S]*?)\\\]/g, (_m, p1) => `$$${String(p1).trim()}$$`)
    .replace(/\\\(([\s\S]*?)\\\)/g, (_m, p1) => `$${String(p1).trim()}$`);
  return md.render(normalized);
});
</script>

<style scoped>
.markdown-view {
  line-height: 1.72;
  font-size: 14px;
  color: #334155;
  overflow-x: auto;
}

.markdown-view :deep(p) {
  margin: 0;
}

.markdown-view :deep(p + p) {
  margin-top: 8px;
}

.markdown-view :deep(ul),
.markdown-view :deep(ol) {
  margin: 8px 0 0;
  padding-left: 20px;
}

.markdown-view :deep(code) {
  background: rgba(148, 163, 184, 0.16);
  border-radius: 6px;
  padding: 1px 5px;
}

.markdown-view :deep(pre) {
  margin: 8px 0 0;
  background: rgba(148, 163, 184, 0.12);
  border-radius: 10px;
  padding: 10px;
  overflow-x: auto;
  max-width: 100%;
}

.markdown-view :deep(blockquote) {
  margin: 8px 0 0;
  border-left: 3px solid rgba(59, 130, 246, 0.42);
  padding-left: 10px;
  color: #475569;
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

.markdown-view.is-dark :deep(.katex) {
  color: #cbd5e1;
}
</style>
