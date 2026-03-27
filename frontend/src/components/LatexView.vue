<template>
  <div class="latex-view" v-html="html"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import katex from 'katex';

const props = defineProps<{
  source: string;
}>();

const html = computed(() => {
  const content = (props.source || '').trim();
  if (!content) {
    return '<span class="latex-empty">暂无内容</span>';
  }

  try {
    return katex.renderToString(content, {
      throwOnError: false,
      displayMode: true,
      trust: true,
      strict: 'ignore',
    });
  } catch {
    return `<pre class="latex-fallback">${escapeHtml(content)}</pre>`;
  }
});

function escapeHtml(input: string) {
  return input
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
}
</script>

<style scoped>
.latex-view {
  overflow-x: auto;
  padding: 8px 2px;
}

:deep(.latex-empty) {
  color: rgba(20, 32, 51, 0.6);
  font-size: 13px;
}

:deep(.latex-fallback) {
  margin: 0;
  white-space: pre-wrap;
  font-size: 13px;
  line-height: 1.55;
}
</style>
