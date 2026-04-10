<template>
  <div class="latex-view">
    <template v-if="jsonQuestion">
      <article class="question-card">
        <header class="question-head">
          <div class="question-stem" v-html="renderRichText(jsonQuestion.stem)" />
        </header>

        <ol v-if="jsonQuestion.options.length" class="question-choices">
          <li v-for="(choice, cIdx) in jsonQuestion.options" :key="`json-c-${cIdx}`">
            <span class="choice-label">{{ choiceLabel(Number(cIdx)) }}.</span>
            <span v-html="renderRichText(choice)" />
          </li>
        </ol>

        <ol v-if="jsonQuestion.subQuestions.length" class="question-parts">
          <li v-for="(part, pIdx) in jsonQuestion.subQuestions" :key="`json-p-${pIdx}`" v-html="renderRichText(part)" />
        </ol>
      </article>
    </template>

    <template v-else-if="questions.length">
      <article
        v-for="(q, idx) in questions"
        :key="`${idx}-${q.stem}`"
        class="question-card"
      >
        <header class="question-head">
          <div class="question-stem" v-html="renderRichText(q.stem)" />
        </header>

        <ol
          v-if="q.choices.length"
          class="question-choices"
        >
          <li v-for="(choice, cIdx) in q.choices" :key="`${idx}-c-${cIdx}`">
            <span class="choice-label">{{ choiceLabel(cIdx) }}.</span>
            <span v-html="renderRichText(choice)" />
          </li>
        </ol>

        <ol v-if="q.parts.length" class="question-parts">
          <li v-for="(part, pIdx) in q.parts" :key="`${idx}-p-${pIdx}`" v-html="renderRichText(part)" />
        </ol>
      </article>
    </template>

    <div v-else class="latex-raw" v-html="renderRichText(normalizedSource, true)" />

    <span v-if="!normalizedSource" class="latex-empty">暂无内容</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import katex from 'katex';

type ParsedQuestion = {
  stem: string;
  choices: string[];
  parts: string[];
};

const props = defineProps<{
  source?: string;
}>();

const normalizedSource = computed(() => (props.source || '').trim());

const jsonQuestion = computed(() => parseQuestionJSON(normalizedSource.value));
const questions = computed(() => parseExamQuestions(normalizedSource.value));

function parseQuestionJSON(input: string) {
  if (!input) {
    return null;
  }

  try {
    const parsed = JSON.parse(input) as any;
    if (!parsed || typeof parsed !== 'object') {
      return null;
    }

    const stem = String(parsed.stem ?? '').trim();
    if (!stem) {
      return null;
    }

    return {
      questionType: String(parsed.question_type ?? '').trim(),
      stem,
      options: Array.isArray(parsed.options) ? parsed.options.map((it: any) => String(it ?? '').trim()).filter(Boolean) : [],
      subQuestions: Array.isArray(parsed.sub_questions)
        ? parsed.sub_questions.map((it: any) => String(it ?? '').trim()).filter(Boolean)
        : [],
    };
  } catch {
    return null;
  }
}

function parseExamQuestions(input: string): ParsedQuestion[] {
  const list: ParsedQuestion[] = [];
  if (!input) {
    return list;
  }

  const questionRe = /\\begin\{question\}(?:\[([^\]]*)\])?([\s\S]*?)\\end\{question\}/g;
  let match: RegExpExecArray | null;

  while ((match = questionRe.exec(input)) !== null) {
    const options = (match[1] || '').trim();
    const body = (match[2] || '').trim();
    //const index = pickQuestionIndex(options, list.length + 1);

    const { contentWithoutChoices, choices } = extractChoices(body);
    const { contentWithoutEnum, parts } = extractEnumerate(contentWithoutChoices);

    list.push({
      stem: cleanupText(contentWithoutEnum),
      choices,
      parts,
    });
  }

  return list;
}

function extractChoices(body: string) {
  const choicesRe = /\\begin\{choices\}(?:\[([^\]]*)\])?([\s\S]*?)\\end\{choices\}/;
  const m = body.match(choicesRe);

  if (!m) {
    return {
      contentWithoutChoices: body,
      choices: [] as string[],
      choiceColumns: 1,
    };
  }

  const options = (m[1] || '').trim();
  const raw = (m[2] || '').trim();
  const itemRe = /\\item\s+([\s\S]*?)(?=(?:\\item\s+)|$)/g;

  const choices: string[] = [];
  let item: RegExpExecArray | null;
  while ((item = itemRe.exec(raw)) !== null) {
    choices.push(cleanupText(item[1] || ''));
  }

  return {
    contentWithoutChoices: cleanupText(body.replace(m[0], '')),
    choices
  };
}

function extractEnumerate(body: string) {
  const enumRe = /\\begin\{enumerate\}([\s\S]*?)\\end\{enumerate\}/;
  const m = body.match(enumRe);

  if (!m) {
    return {
      contentWithoutEnum: body,
      parts: [] as string[],
    };
  }

  const raw = (m[1] || '').trim();
  const itemRe = /\\item\s+([\s\S]*?)(?=(?:\\item\s+)|$)/g;

  const parts: string[] = [];
  let item: RegExpExecArray | null;
  while ((item = itemRe.exec(raw)) !== null) {
    parts.push(cleanupText(item[1] || ''));
  }

  return {
    contentWithoutEnum: cleanupText(body.replace(m[0], '')),
    parts,
  };
}

function cleanupText(text: string) {
  return text
    .replace(/\r/g, '')
    .replace(/^\s+|\s+$/g, '')
    .replace(/\n{3,}/g, '\n\n');
}

function renderRichText(input: string, displayAsBlock = false) {
  const content = cleanupText(input);
  if (!content) {
    return '<span class="latex-empty">暂无内容</span>';
  }

  if (!/[\$]/.test(content) && looksLikeLatexExpression(content)) {
    return renderMath(content, displayAsBlock);
  }

  const paToken = '__EXAM_PA_BLANK__';
  const fillinToken = '__EXAM_FILLIN_BLANK__';
  const normalized = content
    .replace(/\\pa/g, paToken)
    .replace(/\\fillin(?:\[[^\]]*\])?\[[^\]]*\]/g, fillinToken);

  const tokens = normalized.split(/(\$\$[\s\S]+?\$\$|\$[^$\n]+\$)/g).filter(Boolean);
  const htmlParts = tokens.map((token) => {
    if (token.startsWith('$$') && token.endsWith('$$')) {
      return renderMath(token.slice(2, -2), true);
    }

    if (token.startsWith('$') && token.endsWith('$')) {
      return renderMath(token.slice(1, -1), false);
    }

    return applyBlankTokens(escapeHtml(token).replace(/\n/g, '<br/>'), paToken, fillinToken);
  });

  if (displayAsBlock) {
    return `<div class="raw-block">${htmlParts.join('')}</div>`;
  }

  return htmlParts.join('');
}

function applyBlankTokens(input: string, paToken: string, fillinToken: string) {
  return input
    .replaceAll(paToken, '<span class="blank-pa">（&nbsp;&nbsp;&nbsp;&nbsp;）</span>')
    .replaceAll(fillinToken, '<span class="blank-fillin"></span>');
}

function renderMath(expr: string, displayMode: boolean) {
  try {
    return katex.renderToString(expr, {
      throwOnError: false,
      displayMode,
      trust: false,
      strict: 'ignore',
    });
  } catch {
    return `<code>${escapeHtml(expr)}</code>`;
  }
}

function escapeHtml(input: string) {
  return input
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
}

function looksLikeLatexExpression(input: string) {
  const value = input.trim();
  if (!value) {
    return false;
  }
  return /\\[a-zA-Z]+|\^|_|\{.*\}|\frac|\sqrt|\boxed/.test(value);
}

function choiceLabel(index: number) {
  return String.fromCharCode(65 + index);
}

function choiceGridStyle(columns: number) {
  const col = Math.max(1, Math.min(columns || 1, 6));
  return {
    gridTemplateColumns: `repeat(${col}, minmax(0, 1fr))`,
  };
}
</script>

<style scoped>
.latex-view {
  overflow-x: auto;
  padding: 8px 2px;
  display: grid;
  gap: 12px;
}

.question-card {
  border: 1px solid rgba(148, 163, 184, 0.3);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.65);
  padding: 10px 12px;
}

.question-head {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.7;
}

.question-no {
  font-weight: 700;
  color: #334155;
  min-width: 22px;
}

.question-stem {
  flex: 1;
}

.plain-json {
  white-space: normal;
  word-break: break-word;
}

.question-choices {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px 12px;
}

.question-choices li {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  min-width: 0;
}

.choice-label {
  font-weight: 600;
  color: #475569;
}

.question-parts {
  margin: 10px 0 0;
  padding-left: 20px;
  display: grid;
  gap: 8px;
}

.latex-raw {
  line-height: 1.7;
}

.dark .question-card {
  background: transparent;
}

:deep(.blank-fillin) {
  display: inline-block;
  min-width: 4em;
  border-bottom: 1px solid rgba(51, 65, 85, 0.7);
  vertical-align: baseline;
}

:deep(.blank-fillin) {
  min-height: 1.2em;
}

:deep(.latex-empty) {
  color: rgba(20, 32, 51, 0.6);
  font-size: 13px;
}

:deep(.raw-block) {
  white-space: normal;
}
</style>
