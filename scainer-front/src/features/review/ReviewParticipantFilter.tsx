type Props = {
  value: string;
  onChange: (value: string) => void;
};

export default function ReviewParticipantFilter({ value, onChange }: Props) {
  return (
    <label className="contest-head__participant-filter">
      <span className="contest-head__filter-label">Участник</span>
      <input
        type="text"
        className="contest-head__participant-filter-input"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="6B или /.*6B.*/"
        spellCheck={false}
        aria-label="Фильтр по участнику"
      />
    </label>
  );
}
