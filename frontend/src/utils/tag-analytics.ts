import type { ErrorRecord } from '@/types/domain';

export interface TagSummary {
  tag: string;
  recordCount: number;
  totalReviews: number;
  correctCount: number;
  wrongCount: number;
  correctRate: number;
  wrongRate: number;
  avgReviewCount: number;
  avgMastery: number;
  priorityScore: number;
}

export interface TagRelation {
  source: string;
  target: string;
  weight: number;
}

export interface TagGraphData {
  nodes: TagSummary[];
  edges: TagRelation[];
}

export interface TagGraphOptions {
  maxNodes?: number;
  maxEdges?: number;
  allowedTags?: Iterable<string>;
}

export interface RecordWithTagPriority {
  id: number;
  primaryTag: string;
  score: number;
}

function normalizeTagList(tags?: string[]) {
  if (!Array.isArray(tags)) {
    return ['未标记'];
  }
  const cleaned = tags
    .map((tag) => String(tag || '').trim())
    .filter((tag) => tag.length > 0);
  if (!cleaned.length) {
    return ['未标记'];
  }
  return Array.from(new Set(cleaned));
}

export function buildTagSummaries(records: ErrorRecord[]): TagSummary[] {
  const bucket = new Map<string, {
    count: number;
    totalReviews: number;
    totalMastery: number;
    correct: number;
    wrong: number;
  }>();

  records.forEach((record) => {
    const reviewCount = Number(record.reviewCount || 0);
    const mastery = Number(record.masteryLevel || 0);
    const isCorrect = record.lastReviewResult === 'correct';
    const isWrong = record.lastReviewResult === 'wrong';

    normalizeTagList(record.questionTags).forEach((tag) => {
      const current = bucket.get(tag) || {
        count: 0,
        totalReviews: 0,
        totalMastery: 0,
        correct: 0,
        wrong: 0,
      };
      current.count += 1;
      current.totalReviews += reviewCount;
      current.totalMastery += mastery;
      if (isCorrect) {
        current.correct += 1;
      }
      if (isWrong) {
        current.wrong += 1;
      }
      bucket.set(tag, current);
    });
  });

  return [...bucket.entries()]
    .map(([tag, value]) => {
      const correctRate = value.count ? Math.round((value.correct / value.count) * 100) : 0;
      const wrongRate = value.count ? Math.round((value.wrong / value.count) * 100) : 0;
      const avgReviewCount = value.count ? Math.round((value.totalReviews / value.count) * 10) / 10 : 0;
      const avgMastery = value.count ? Math.round(value.totalMastery / value.count) : 0;
      // Higher wrong rate + higher review load + lower mastery => higher review priority.
      const priorityScore = Math.round((wrongRate * 0.55 + avgReviewCount * 10 * 0.25 + (100 - avgMastery) * 0.2) * 10) / 10;
      return {
        tag,
        recordCount: value.count,
        totalReviews: value.totalReviews,
        correctCount: value.correct,
        wrongCount: value.wrong,
        correctRate,
        wrongRate,
        avgReviewCount,
        avgMastery,
        priorityScore,
      };
    })
    .sort((a, b) => b.priorityScore - a.priorityScore || b.recordCount - a.recordCount || a.tag.localeCompare(b.tag));
}

export function buildTagRelations(records: ErrorRecord[], maxEdges = 24, allowedTags?: Iterable<string>): TagRelation[] {
  const pairMap = new Map<string, number>();
  const allowed = allowedTags ? new Set(Array.from(allowedTags, (tag) => String(tag || '').trim()).filter((tag) => tag.length > 0)) : null;

  records.forEach((record) => {
    const tags = normalizeTagList(record.questionTags);
    const filteredTags = allowed ? tags.filter((tag) => allowed.has(tag)) : tags;
    for (let i = 0; i < filteredTags.length; i += 1) {
      for (let j = i + 1; j < filteredTags.length; j += 1) {
        const source = filteredTags[i];
        const target = filteredTags[j];
        const key = source < target ? `${source}__${target}` : `${target}__${source}`;
        pairMap.set(key, (pairMap.get(key) || 0) + 1);
      }
    }
  });

  return [...pairMap.entries()]
    .map(([key, weight]) => {
      const [source, target] = key.split('__');
      return { source, target, weight };
    })
    .sort((a, b) => b.weight - a.weight)
    .slice(0, maxEdges);
}

export function buildTagGraphData(records: ErrorRecord[], options: TagGraphOptions = {}): TagGraphData {
  const maxNodes = options.maxNodes ?? 10;
  const maxEdges = options.maxEdges ?? 24;
  const allowedTags = options.allowedTags;

  return {
    nodes: buildTagSummaries(records).slice(0, maxNodes),
    edges: buildTagRelations(records, maxEdges, allowedTags),
  };
}

export function topPriorityTag(records: ErrorRecord[]): TagSummary | null {
  const list = buildTagSummaries(records);
  return list.length ? list[0] : null;
}

export function scoreRecordByTags(record: ErrorRecord, tagPriority: Map<string, number>): RecordWithTagPriority {
  const tags = normalizeTagList(record.questionTags);
  let bestTag = tags[0] || '未标记';
  let bestScore = -1;
  tags.forEach((tag) => {
    const score = tagPriority.get(tag) ?? 0;
    if (score > bestScore) {
      bestScore = score;
      bestTag = tag;
    }
  });
  return {
    id: record.id,
    primaryTag: bestTag,
    score: Math.max(bestScore, 0),
  };
}

export function summarizeTags(tags?: string[], limit = 3) {
  const list = normalizeTagList(tags);
  return list.slice(0, limit);
}
