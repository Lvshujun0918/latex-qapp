export type SyncStatus = 'pending' | 'syncing' | 'synced' | 'conflict' | 'failed';

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
  difficulty: number;
  title?: string;
  latexSource: string;
  latexAnswer?: string;
  questionTags?: string[];
  latexVersion: number;
  latexRenderStatus: 'pending' | 'ok' | 'failed';
  mistakeReason?: string;
  masteryLevel: number;
  reviewCount: number;
  syncStatus: SyncStatus;
  localVersion: number;
  serverVersion: number;
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
  latexQuestion: string;
  latexAnswer: string;
  tags: string[];
  subject?: string;
  title?: string;
}

export interface PDFJob {
  id: number;
  jobNo: string;
  status: 'queued' | 'running' | 'success' | 'partial' | 'failed';
  progress: number;
  pdfFileUrl?: string;
}
