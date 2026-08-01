export type ReviewVerdictFilter = {
  active: boolean;
  verdicts: string[];
};

export const DEFAULT_VERDICT_FILTER: ReviewVerdictFilter = {
  active: true,
  verdicts: ["PR"],
};

/** Порядок в выпадашке фильтра (OK в API отображается как AC). */
export const REVIEW_VERDICT_OPTIONS = [
  "PR",
  "OK",
  "RJ",
  "WA",
  "RT",
  "IG",
  "PE",
  "TL",
  "ML",
  "CE",
  "CF",
  "DQ",
] as const;

export type ReviewVerdictOption = (typeof REVIEW_VERDICT_OPTIONS)[number];

export type ReviewParticipantFilter = {
  active: boolean;
  query: string;
};

export const DEFAULT_PARTICIPANT_FILTER: ReviewParticipantFilter = {
  active: true,
  query: "",
};

export type ReviewFiltersInput = {
  verdictFilter: ReviewVerdictFilter;
  participantFilter: ReviewParticipantFilter;
};
