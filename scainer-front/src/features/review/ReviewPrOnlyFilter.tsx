type Props = {
  checked: boolean;
  onChange: (value: boolean) => void;
};

export default function ReviewPrOnlyFilter({ checked, onChange }: Props) {
  return (
    <label className="contest-head__pr-filter">
      <input
        type="checkbox"
        className="contest-head__pr-filter-input"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
      />
      <span className="contest-head__pr-filter-label">PR only</span>
    </label>
  );
}
