export interface User {
  id: number;
  username: string;
  displayName: string;
}

export interface ErrorRecord {
  id: number;
  userId: number;
  subject: string;
  questionType?: string;
  title?: string;
  latexSource: string;
  latexAnswer?: string;
  questionTags?: string[];
  latexVersion: number;
  latexRenderStatus: 'pending' | 'ok' | 'failed';
  mistakeReason?: string;
  masteryLevel: number;
  reviewCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface AIAnalysisResult {
  steps: string[];
  keyPoints: string[];
  suggestions: string[];
  summary: string;
}

export interface VisionLatexDraft {
  questionJson?: {
    question_type?: string;
    stem: string;
    options?: string[];
    sub_questions?: string[];
    latex_answer?: string;
  };
  latexSource: string;
  latexQuestion: string;
  latexAnswer: string;
  latexSolution?: string;
  tags: string[];
  questionType?: string;
  subject?: string;
  title?: string;
}

export interface VisionStreamEvent {
  stage: 'classify' | 'latex' | 'tags' | 'solve' | 'final' | 'error';
  subject?: string;
  title?: string;
  question_type?: string;
  question_json?: {
    question_type?: string;
    stem: string;
    options?: string[];
    sub_questions?: string[];
    latex_answer?: string;
  };
  latex_source?: string;
  latex_answer?: string;
  latex_solution?: string;
  tags?: string[];
  raw_content?: string;
  agent_trace?: string[];
  done?: boolean;
  error?: string;
}

export interface PDFJob {
  id: number;
  jobNo: string;
  status: 'queued' | 'running' | 'success' | 'partial' | 'failed';
  progress: number;
  pdfFileUrl?: string;
}
