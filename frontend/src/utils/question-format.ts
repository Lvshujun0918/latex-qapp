export type NormalizedQuestionJson = {
  questionType: string;
  stem: string;
  options: string[];
  subQuestions: string[];
};

const LEADING_PREFIX_PATTERNS = [
  /^\s*第\s*\d+\s*题\s*/,
  /^\s*题目\s*\d+\s*/,
  /^\s*[（(]?\d+[)）]?\s*[.、．)]\s*/,
  /^\s*[（(]?[A-Ha-h][)）]?\s*[.、．)]\s*/,
  /^\s*[（(]\d+[)）]\s*/,
  /^\s*[（(][A-Ha-h][)）]\s*/,
];

export function cleanQuestionText(text: string): string {
  let value = String(text || '').trim();
  if (!value) {
    return '';
  }

  for (let i = 0; i < 3; i += 1) {
    const before = value;
    for (const pattern of LEADING_PREFIX_PATTERNS) {
      value = value.replace(pattern, '').trim();
    }
    if (value === before) {
      break;
    }
  }

  return value;
}

export function cleanChoiceText(text: string): string {
  return cleanQuestionText(text);
}

export function cleanPartText(text: string): string {
  return cleanQuestionText(text);
}

export function normalizeQuestionTypeLabel(input?: string): string {
  const value = String(input || '').trim().toLowerCase();
  if (['choice', '选择', '选择题', 'single_choice', 'multiple_choice'].includes(value)) return '选择';
  if (['fill_blank', '填空', '填空题'].includes(value)) return '填空';
  if (['essay', '解答', '解答题', 'subjective', '大题'].includes(value)) return '解答';
  if (!value || value === 'unknown' || value === '未知') return '未知';
  return String(input || '未知');
}

export function normalizeSubjectLabel(input?: string): string {
  const value = String(input || '').trim().toLowerCase();
  if (value === 'math' || value === '数学') return '数学';
  if (value === 'physics' || value === '物理') return '物理';
  if (value === 'chemistry' || value === '化学') return '化学';
  if (value === 'biology' || value === '生物') return '生物';
  if (!value || value === 'unknown' || value === '未知') return '未知';
  return String(input || '未知');
}

export function assembleLatexFromQuestionJson(questionJson: any): string {
  if (!questionJson || typeof questionJson !== 'object') {
    return '';
  }

  const questionType = normalizeQuestionTypeLabel(questionJson.question_type);
  const stem = cleanQuestionText(questionJson.stem ?? '');
  if (!stem) {
    return '';
  }

  if (questionType === '选择') {
    const options = Array.isArray(questionJson.options)
      ? questionJson.options.map((item: any) => cleanChoiceText(String(item ?? ''))).filter(Boolean)
      : [];
    if (!options.length) {
      return stem;
    }
    return `${stem}\n\\begin{choices}\n${options.map((opt: string) => `\\item ${opt}`).join('\n')}\n\\end{choices}`;
  }

  if (questionType === '解答') {
    const subQuestions = Array.isArray(questionJson.sub_questions)
      ? questionJson.sub_questions.map((item: any) => cleanPartText(String(item ?? ''))).filter(Boolean)
      : [];
    if (!subQuestions.length) {
      return stem;
    }
    return `${stem}\n\\begin{enumerate}\n${subQuestions.map((item: string) => `\\item ${item}`).join('\n')}\n\\end{enumerate}`;
  }

  return stem;
}
