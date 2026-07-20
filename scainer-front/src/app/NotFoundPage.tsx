import { Link } from "react-router-dom";

export default function NotFoundPage() {
  return (
    <div className="page-center">
      <h1>404 — Страница не найдена</h1>
      <Link to="/" className="btn">
        Вернуться в список контестов
      </Link>
    </div>
  );
}
