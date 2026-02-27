'use client';

import { useState, useEffect, useCallback, FormEvent } from 'react';

// Типизация данных хакатона
interface Hackathon {
  id: string;
  title: string;
  date: string;
  format: string;
  city: string;
  ageLimit: string;
  link: string;
  status: string;
}

const loadingMessages = [
  "Гуглим интернет...",
  "Читаем сайты...",
  "Анализируем через ИИ..."
];

export default function Home() {
  const [query, setQuery] = useState('');
  const [hackathons, setHackathons] = useState<Hackathon[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingText, setLoadingText] = useState('');
  const [error, setError] = useState<string | null>(null);

  // Функция загрузки данных с бэкенда
  const fetchHackathons = useCallback(async (searchQuery: string = '') => {
    setLoading(true);
    setError(null);
    let intervalId: NodeJS.Timeout | null = null;

    if (searchQuery) {
      setLoadingText(loadingMessages[0]);
      let index = 0;
      intervalId = setInterval(() => {
        index = (index + 1) % loadingMessages.length;
        setLoadingText(loadingMessages[index]);
      }, 1500); // Меняем текст каждые 1.5 секунды
    }

    try {
      const controller = new AbortController();
      // Увеличиваем таймаут для AI агента
      const timeoutId = setTimeout(() => controller.abort(), 20000);

      const endpoint = searchQuery
        ? `http://localhost:8080/api/search?q=${encodeURIComponent(searchQuery)}`
        : `http://localhost:8080/api/hackathons`;

      const res = await fetch(endpoint, {
        signal: controller.signal,
        headers: {
          'Accept': 'application/json',
        }
      });

      clearTimeout(timeoutId);

      if (!res.ok) {
        throw new Error(`Ошибка HTTP: ${res.status}`);
      }

      const data = await res.json();
      setHackathons(data || []);
    } catch (err: any) {
      console.error('Ошибка при загрузке хакатонов:', err);
      if (err.name !== 'AbortError') {
        setError('Не удалось загрузить данные серверов. Проверьте запущен ли бэкенд.');
      }
    } finally {
      if (intervalId) clearInterval(intervalId);
      setLoading(false);
      setLoadingText('');
    }
  }, []);

  // Первоначальная загрузка
  useEffect(() => {
    fetchHackathons();
  }, [fetchHackathons]);

  // Обработчик отправки формы
  const handleSearch = (e: FormEvent) => {
    e.preventDefault();
    fetchHackathons(query);
  };

  return (
    <main className="min-h-screen bg-[#050510] relative overflow-hidden text-zinc-300 font-mono selection:bg-[#00ff9d] selection:text-black">
      {/* Background glowing effects */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[500px] bg-[#00ff9d]/5 rounded-full blur-[120px] pointer-events-none" />
      <div className="absolute bottom-0 right-0 w-[600px] h-[400px] bg-blue-500/5 rounded-full blur-[100px] pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 space-y-16">

        {/* Header */}
        <header className="text-center space-y-6">
          <div className="inline-block relative">
            <h1 className="text-6xl md:text-8xl font-black tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-[#00ff9d] via-[#00f0ff] to-[#00ff9d] animate-gradient-x drop-shadow-[0_0_20px_rgba(0,255,157,0.3)]">
              HACKFLOW
            </h1>
          </div>
          <p className="max-w-2xl mx-auto text-lg md:text-xl text-zinc-400">
            Платформа для поиска IT-эвентов. <br className="hidden sm:block" />
            <span className="text-white border-b-2 border-[#00ff9d]/30 pb-1">Никаких галлюцинаций — только проверенные хакатоны.</span>
          </p>
        </header>

        {/* Search Form */}
        <section className="max-w-3xl mx-auto">
          <form onSubmit={handleSearch} className="relative group">
            {/* Glow effect behind the input */}
            <div className="absolute -inset-1 bg-gradient-to-r from-[#00ff9d] to-[#00f0ff] rounded-2xl opacity-20 blur transition duration-500 group-hover:opacity-40" />

            <div className="relative flex flex-col sm:flex-row gap-3 p-2 bg-[#0a0a16]/80 backdrop-blur-xl border border-white/10 rounded-2xl shadow-2xl">
              <div className="relative flex-1 flex items-center">
                <div className="absolute left-4 text-zinc-500">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
                    <path fillRule="evenodd" d="M10.5 3.75a6.75 6.75 0 100 13.5 6.75 6.75 0 000-13.5zM2.25 10.5a8.25 8.25 0 1114.59 5.28l4.69 4.69a.75.75 0 11-1.06 1.06l-4.69-4.69A8.25 8.25 0 012.25 10.5z" clipRule="evenodd" />
                  </svg>
                </div>
                <input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Поиск по городу или названию..."
                  className="w-full bg-transparent border-none outline-none pl-12 pr-4 py-3 text-lg text-white placeholder-zinc-500 rounded-xl focus:ring-0"
                />
              </div>
              <button
                type="submit"
                disabled={loading}
                className="min-w-[140px] px-8 py-3 bg-gradient-to-r from-[#00ff9d] to-[#00cc7d] hover:from-[#00ff9d] hover:to-[#00ff9d] text-black font-bold uppercase tracking-widest text-sm rounded-xl transition-all duration-300 disabled:opacity-80 disabled:cursor-not-allowed shadow-[0_0_20px_rgba(0,255,157,0.2)] hover:shadow-[0_0_30px_rgba(0,255,157,0.4)] active:scale-95"
              >
                {loading && query ? (
                  <div className="flex items-center justify-center gap-2 animate-pulse">
                    <svg className="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span className="text-xs">{loadingText}</span>
                  </div>
                ) : loading ? (
                  <svg className="animate-spin h-5 w-5 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                ) : (
                  "Найти"
                )}
              </button>
            </div>
          </form>
        </section>

        {/* Error Message */}
        {error && (
          <div className="max-w-3xl mx-auto p-4 bg-red-500/10 border border-red-500/20 rounded-xl flex items-center gap-3 text-red-400">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-6 h-6 shrink-0">
              <path fillRule="evenodd" d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003zM12 8.25a.75.75 0 01.75.75v3.75a.75.75 0 01-1.5 0V9a.75.75 0 01.75-.75zm0 8.25a.75.75 0 100-1.5.75.75 0 000 1.5z" clipRule="evenodd" />
            </svg>
            <p className="font-medium">{error}</p>
          </div>
        )}

        {/* Hackathons Grid */}
        <section>
          {loading && hackathons.length === 0 ? (
            <div className="flex justify-center items-center py-32">
              <div className="relative w-16 h-16">
                <div className="absolute inset-0 rounded-full border-t-2 border-[#00ff9d] animate-spin"></div>
                <div className="absolute inset-2 rounded-full border-t-2 border-[#00f0ff] animate-spin-slow"></div>
              </div>
            </div>
          ) : hackathons.length === 0 && !error ? (
            <div className="flex flex-col justify-center items-center py-20 bg-[#0a0a16]/40 backdrop-blur-sm border border-white/5 rounded-3xl">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-16 h-16 text-zinc-600 mb-4">
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.182 16.318A4.486 4.486 0 0012.016 15a4.486 4.486 0 00-3.198 1.318M21 12a9 9 0 11-18 0 9 9 0 0118 0zM9.75 9.75c0 .414-.168.75-.375.75S9 10.164 9 9.75 9.168 9 9.375 9s.375.336.375.75zm3.625 0c0 .414-.168.75-.375.75s-.375-.336-.375-.75.168-.75.375-.75.375.336.375.75z" />
              </svg>
              <h3 className="text-2xl font-bold text-white mb-2">Хакатоны не найдены</h3>
              <p className="text-zinc-500 font-sans">Попробуйте изменить параметры поиска или ключевые слова</p>
            </div>
          ) : (
            <div className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 transition-opacity duration-500 ${loading ? 'opacity-30 pointer-events-none' : 'opacity-100'}`}>
              {hackathons.map((h, index) => (
                <div
                  key={h.id || `ai-result-${index}`}
                  className="group relative flex flex-col bg-[#0f0f1d]/80 backdrop-blur-md rounded-2xl border border-white/5 hover:border-[#00ff9d]/30 overflow-hidden transition-all duration-300 hover:shadow-[0_8px_30px_rgba(0,255,157,0.08)] hover:-translate-y-1"
                >
                  {/* Top Highlight Line */}
                  <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-[#00ff9d] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                  <div className="p-6 flex-1 flex flex-col">
                    {/* Format and Date */}
                    <div className="flex flex-wrap gap-2 items-center mb-6">
                      <span className={`px-2.5 py-1 text-xs font-bold uppercase tracking-wider rounded-md border backdrop-blur-sm
                        ${h.format.includes('Онлайн')
                          ? 'bg-blue-500/10 text-blue-400 border-blue-500/20'
                          : 'bg-purple-500/10 text-purple-400 border-purple-500/20'
                        }`}
                      >
                        {h.format}
                      </span>
                      {h.status === 'DEAD' && (
                        <span className="px-2.5 py-1 text-xs font-bold uppercase tracking-wider rounded-md border bg-red-500/10 text-red-400 border-red-500/20 backdrop-blur-sm">
                          Завершено
                        </span>
                      )}
                      <span className="inline-flex items-center gap-1.5 text-xs text-zinc-400 font-sans bg-white/5 px-2.5 py-1 rounded-md border border-white/5 ml-auto">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-3.5 h-3.5">
                          <path fillRule="evenodd" d="M5.75 2a.75.75 0 01.75.75V4h7V2.75a.75.75 0 011.5 0V4h.25A2.75 2.75 0 0118 6.75v8.5A2.75 2.75 0 0115.25 18H4.75A2.75 2.75 0 012 15.25v-8.5A2.75 2.75 0 014.75 4H5V2.75A.75.75 0 015.75 2zm-1 5.5c-.69 0-1.25.56-1.25 1.25v6.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-6.5c0-.69-.56-1.25-1.25-1.25H4.75z" clipRule="evenodd" />
                        </svg>
                        {h.date}
                      </span>
                    </div>

                    {/* Title */}
                    <h2 className="text-2xl font-bold text-white mb-4 leading-tight group-hover:text-[#00ff9d] transition-colors">
                      {h.title}
                    </h2>

                    {/* Details */}
                    <div className="space-y-3 font-sans text-sm mb-8 mt-auto">
                      <div className="flex items-center gap-3 text-zinc-400">
                        <div className="w-7 h-7 rounded-full bg-white/5 flex items-center justify-center shrink-0">
                          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4 text-zinc-500">
                            <path fillRule="evenodd" d="M9.69 18.933l.003.001C9.89 19.02 10 19 10 19s.11.02.308-.066l.002-.001.006-.003.018-.008a5.741 5.741 0 00.281-.14c.186-.096.446-.24.757-.433.62-.384 1.445-.966 2.274-1.765C15.302 14.988 17 12.493 17 9A7 7 0 103 9c0 3.492 1.698 5.988 3.355 7.584a13.731 13.731 0 002.273 1.765 11.842 11.842 0 00.976.544l.062.029.018.008.006.003zM10 11.25a2.25 2.25 0 100-4.5 2.25 2.25 0 000 4.5z" clipRule="evenodd" />
                          </svg>
                        </div>
                        <span className="truncate">{h.city}</span>
                      </div>
                      <div className="flex items-center gap-3 text-zinc-400">
                        <div className="w-7 h-7 rounded-full bg-white/5 flex items-center justify-center shrink-0">
                          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4 text-zinc-500">
                            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-5.5-2.5a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0zM10 12a5.99 5.99 0 00-4.793 2.39A6.483 6.483 0 0010 16.5a6.483 6.483 0 004.793-2.11A5.99 5.99 0 0010 12z" clipRule="evenodd" />
                          </svg>
                        </div>
                        <span className="truncate">Возраст: {h.ageLimit}</span>
                      </div>
                    </div>
                  </div>

                  {/* Action Button */}
                  <div className="p-4 pt-0">
                    <a
                      href={h.link}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center justify-center gap-2 w-full py-3 px-4 bg-white/5 hover:bg-[#00ff9d] text-white hover:text-black font-semibold font-sans rounded-xl transition-colors duration-300 border border-white/5 hover:border-transparent group/btn"
                    >
                      Подробнее
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4 translate-x-0 group-hover/btn:translate-x-1 transition-transform">
                        <path fillRule="evenodd" d="M5.22 14.78a.75.75 0 001.06 0l7.22-7.22v5.69a.75.75 0 001.5 0v-7.5a.75.75 0 00-.75-.75h-7.5a.75.75 0 000 1.5h5.69l-7.22 7.22a.75.75 0 000 1.06z" clipRule="evenodd" />
                      </svg>
                    </a>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>

      {/* Tailwind Custom Keyframes Extension for local use without config */}
      <style dangerouslySetInnerHTML={{
        __html: `
        @keyframes gradient-x {
          0%, 100% {
            background-position: 0% 50%;
          }
          50% {
            background-position: 100% 50%;
          }
        }
        .animate-gradient-x {
          background-size: 200% 200%;
          animation: gradient-x 3s ease infinite;
        }
        .animate-spin-slow {
          animation: spin 3s linear infinite;
        }
      `}} />
    </main>
  );
}
