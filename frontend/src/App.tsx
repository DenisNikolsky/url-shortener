import { useState } from "react";

function App() {
  const [url, setUrl] = useState("");
  const [shortUrl, setShortUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    setError("");
    setShortUrl("");

    if (!url.trim()) {
      setError("Введите URL");
      return;
    }

    try {
      setLoading(true);

      const response = await fetch("http://localhost:8080/urls", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          url: url.trim(),
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Не удалось сократить URL");
      }

      setShortUrl(
          `http://localhost:8080/${data.short_code}`
      );
    } catch (err) {
      setError(
          err instanceof Error
              ? err.message
              : "Произошла ошибка"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(shortUrl);
  };

  return (
      <div className="app">
        <main className="container">
          <div className="logo">🔗</div>

          <h1>URL Shortener</h1>

          <p className="subtitle">
            Сокращай длинные ссылки быстро и удобно
          </p>

          <form onSubmit={handleSubmit} className="url-form">
            <input
                type="url"
                placeholder="https://example.com/very/long/url"
                value={url}
                onChange={(event) => setUrl(event.target.value)}
                disabled={loading}
            />

            <button type="submit" disabled={loading}>
              {loading ? "Сокращаем..." : "Сократить"}
            </button>
          </form>

          {error && (
              <div className="error">
                {error}
              </div>
          )}

          {shortUrl && (
              <div className="result">
                <span>Ваша короткая ссылка:</span>

                <div className="result-row">
                  <a
                      href={shortUrl}
                      target="_blank"
                      rel="noreferrer"
                  >
                    {shortUrl}
                  </a>

                  <button
                      type="button"
                      onClick={handleCopy}
                  >
                    Копировать
                  </button>
                </div>
              </div>
          )}

          <footer>
            Go · Echo · PostgreSQL
          </footer>
        </main>
      </div>
  );
}

export default App;