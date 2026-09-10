export function Pagination({
  page,
  pageSize,
  total,
  onPage,
}: {
  page: number;
  pageSize: number;
  total: number;
  onPage: (page: number) => void;
}) {
  const lastPage = Math.max(1, Math.ceil(total / pageSize));
  const from = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, total);

  return (
    <div className="pager">
      <span className="muted">
        {from}–{to} / {total}
      </span>
      <div className="pager__buttons">
        <button
          type="button"
          className="btn btn--ghost"
          onClick={() => onPage(page - 1)}
          disabled={page <= 1}
        >
          Előző
        </button>
        <span className="pager__page">
          {page}. oldal / {lastPage}
        </span>
        <button
          type="button"
          className="btn btn--ghost"
          onClick={() => onPage(page + 1)}
          disabled={page >= lastPage}
        >
          Következő
        </button>
      </div>
    </div>
  );
}
