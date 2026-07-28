type Props = {
  onClick: () => void;
  className?: string;
};

export default function ReviewShowAllSubmissions({ onClick, className }: Props) {
  const rootClass = ["review-show-all", className].filter(Boolean).join(" ");

  return (
    <button type="button" className={rootClass} onClick={onClick}>
      <span className="review-show-all__rule" aria-hidden />
      <span className="review-show-all__label">Показать все посылки</span>
      <span className="review-show-all__rule" aria-hidden />
    </button>
  );
}
