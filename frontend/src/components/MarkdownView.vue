<template>
  <div class="markdown-view" :class="{ 'is-dark': resolvedTheme === 'dark' }" v-html="html" />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import MarkdownIt from 'markdown-it';
import { useTheme } from '@/composables/useTheme';

const props = defineProps<{
  source?: string;
}>();
const { resolvedTheme } = useTheme();

const md = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
});

const html = computed(() => {
  const input = (props.source || '').trim();
  if (!input) {
    return '<p>暂无解析</p>';
  }
  return md.render(input);
});
</script>

<style scoped>
.markdown-view {
  line-height: 1.72;
  font-size: 14px;
  color: #334155;
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
}

.markdown-view :deep(blockquote) {
  margin: 8px 0 0;
  border-left: 3px solid rgba(59, 130, 246, 0.42);
  padding-left: 10px;
  color: #475569;
}

.markdown-view.is-dark {
  color: #cbd5e1;
}
</style>
