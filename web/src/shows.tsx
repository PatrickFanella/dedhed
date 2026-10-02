import { useEffect, useState } from 'react';
import { clock, monthGrid, MONTHS, type DateItem, type DayCount, type SearchHit, type YearCount } from './types';

async function get<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) { const data = await res.json().catch(() => null); throw new Error(data?.error ?? 'The show index is unavailable'); }
  return res.json() as Promise<T>;
}

export type Route = { view: 'home' } | { view: 'calendar'; year: number | null } | { view: 'date'; date: string };
export function parseRoute(hash: string): Route {
  const parts = hash.replace(/^#\/?/, '').split('/').filter(Boolean);
  if (parts[0] === 'calendar') { const y = Number(parts[1]); return { view: 'calendar', year: parts[1] && Number.isInteger(y) ? y : null }; }
  if (parts[0] === 'date' && /^\d{4}-\d{2}-\d{2}$/.test(parts[1] ?? '')) return { view: 'date', date: parts[1] };
  return { view: 'home' };
}
export function useRoute(): Route {
  const [route, setRoute] = useState<Route>(() => parseRoute(window.location.hash));
  useEffect(() => { const on = () => setRoute(parseRoute(window.location.hash)); window.addEventListener('hashchange', on); return () => window.removeEventListener('hashchange', on); }, []);
  return route;
}

/** The value most tapes agree on; ties go to the earliest in list order. */
export function commonest(values: string[]): string {
  const counts = new Map<string, number>();
  for (const v of values) if (v) counts.set(v, (counts.get(v) ?? 0) + 1);
  let best = '', n = 0;
  for (const [v, c] of counts) if (c > n) { best = v; n = c; }
  return best;
}
function stock(items: number): string { return items >= 8 ? 'stock-red' : items >= 4 ? 'stock-yellow' : items >= 2 ? 'stock-sky' : ''; }
function kindTab(kind: string): string { return kind === 'soundboard' ? 'tab gold' : kind === 'matrix' ? 'tab chrome' : kind === 'synthetic' ? 'tab chrome' : 'tab'; }
function kindLabel(kind: string): string { return ({ soundboard: 'SBD', audience: 'AUD', matrix: 'MTX', synthetic: 'TEST' } as Record<string, string>)[kind] ?? '?'; }

/** The strip of shows that happened on today's month and day, across every year. */
export function OnThisDay() {
  const [data, setData] = useState<{ month: number; day: number; dates: DayCount[] } | null>(null);
  const [error, setError] = useState('');
  useEffect(() => { let active = true; get<{ month: number; day: number; dates: DayCount[] }>('/api/on-this-day').then(d => { if (active) setData(d); }).catch(e => { if (active) setError(e.message); }); return () => { active = false; }; }, []);
  if (error) return <section className="today" aria-label="On this day"><p className="eyebrow">On this day</p><p className="status">{error}</p></section>;
  if (!data) return <section className="today" aria-label="On this day"><p className="eyebrow">On this day</p><p className="status" role="status">Checking the calendar…</p></section>;
  return <section className="today" aria-label="On this day">
    <div className="today-head"><p className="eyebrow">On this day · {MONTHS[data.month - 1]} {data.day}</p><a href="#/calendar" className="eyebrow link">Open the calendar →</a></div>
    {data.dates.length === 0 ? <p className="shelf-empty">No tapes in the collection for this date.</p> : <ul className="today-list">{data.dates.map(d => <li key={d.date}><a className="spine" href={`#/date/${d.date}`}><span className="tab">{d.date.slice(0, 4)}</span><span className={`label ${stock(d.items)}`}><span className="band">{d.date}</span><span className="venue">{d.venue || 'Venue unrecorded'}</span></span><span className="end">{d.items} {d.items === 1 ? 'tape' : 'tapes'}</span></a></li>)}</ul>}
  </section>;
}

/** Year strip plus twelve month grids. Days with tapes are coloured by how many. */
export function CalendarView({ year }: { year: number | null }) {
  const [years, setYears] = useState<YearCount[]>([]);
  const [days, setDays] = useState<DayCount[]>([]);
  const [error, setError] = useState(''); const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState(''); const [hits, setHits] = useState<SearchHit[] | null>(null); const [total, setTotal] = useState(0); const [page, setPage] = useState(1);
  const chosen = year ?? (years.length ? years[0].year : null);
  useEffect(() => { let active = true; setLoading(true); setError('');
    get<{ years: YearCount[]; days?: DayCount[] }>(`/api/calendar${year ? `?year=${year}` : ''}`).then(d => { if (!active) return; setYears(d.years); if (year) setDays(d.days ?? []); else if (d.years.length) window.location.hash = `#/calendar/${d.years[0].year}`; }).catch(e => { if (active) setError(e.message); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; }; }, [year]);
  useEffect(() => { if (!query.trim()) { setHits(null); return; } let active = true; const t = window.setTimeout(() => {
    get<{ dates: SearchHit[]; total: number }>(`/api/shows?${new URLSearchParams({ q: query, page: String(page) })}`).then(d => { if (active) { setHits(d.dates); setTotal(d.total); } }).catch(e => { if (active) setError(e.message); }); }, 250);
    return () => { active = false; window.clearTimeout(t); }; }, [query, page]);
  const byDate = new Map(days.map(d => [d.date, d]));
  return <section className="calendar" aria-label="Calendar">
    <div className="library-head"><h2>The calendar <small>{years.reduce((n, y) => n + y.items, 0).toLocaleString()} tapes · {years.reduce((n, y) => n + y.dates, 0).toLocaleString()} dates</small></h2>
      <div className="filters"><label className="field"><span>Search every tape</span><input type="search" placeholder="Song, venue, city, taper…" value={query} onChange={e => { setQuery(e.target.value); setPage(1); }} /></label></div></div>
    {error && <p className="status" role="alert">{error}</p>}
    {hits ? <div className="results">
      <p className="eyebrow">{total.toLocaleString()} {total === 1 ? 'date matches' : 'dates match'}</p>
      {hits.length === 0 && <p className="shelf-empty">Nothing in the collection matches. Try a song title, a venue or a city.</p>}
      {hits.map(h => <article className="hit" key={h.date}><a className="hit-date" href={`#/date/${h.date}`}><strong>{h.date}</strong> {h.venue}</a><ul>{h.items.map(it => <li key={it.id}><span className="sticker">{kindLabel(it.kind)}</span> {it.title || it.id}{it.songs && <small> · {it.songs.length > 140 ? it.songs.slice(0, 140) + '…' : it.songs}</small>}</li>)}</ul></article>)}
      {total > 20 && <nav className="pager" aria-label="Result pages"><button className="key ghost" disabled={page === 1} onClick={() => setPage(page - 1)}>← Previous</button><span>Page {page} of {Math.ceil(total / 20)}</span><button className="key ghost" disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>Next →</button></nav>}
    </div> : <>
      <ul className="years" aria-label="Years">{years.map(y => <li key={y.year}><a href={`#/calendar/${y.year}`} aria-current={y.year === chosen ? 'page' : undefined}>{y.year}<small>{y.dates}</small></a></li>)}</ul>
      {loading && <p className="status" role="status">Pulling the calendar…</p>}
      {!loading && years.length === 0 && !error && <p className="shelf-empty">The show index is empty. Run the inventory import to fill the calendar.</p>}
      {chosen && !loading && <div className="months">{MONTHS.map((name, m) => { const grid = monthGrid(chosen, m + 1); const count = days.filter(d => d.date.slice(5, 7) === String(m + 1).padStart(2, '0')).length; return <div className="month" key={name}><h3>{name} <small>{count ? `${count} ${count === 1 ? 'date' : 'dates'}` : ''}</small></h3><div className="grid" role="grid" aria-label={`${name} ${chosen}`}>{['M', 'T', 'W', 'T', 'F', 'S', 'S'].map((d, i) => <span className="dow" key={i} aria-hidden="true">{d}</span>)}{grid.map((cell, i) => { if (!cell) return <span className="blank" key={i} />; const hit = byDate.get(cell); return hit ? <a className={`day has ${stock(hit.items)}`} href={`#/date/${cell}`} key={i} title={`${hit.venue} · ${hit.items} ${hit.items === 1 ? 'tape' : 'tapes'}`}>{Number(cell.slice(8))}</a> : <span className="day" key={i}>{Number(cell.slice(8))}</span>; })}</div></div>; })}</div>}
    </>}
  </section>;
}

/** Every recording the collection holds for one date, playable ones first. */
export function DateView({ date, onPlay, busy }: { date: string; onPlay: (id: string) => void; busy: string }) {
  const [items, setItems] = useState<DateItem[] | null>(null); const [error, setError] = useState('');
  useEffect(() => { let active = true; setItems(null); setError(''); get<{ items: DateItem[] }>(`/api/dates/${date}`).then(d => { if (active) setItems(d.items); }).catch(e => { if (active) setError(e.message); }); return () => { active = false; }; }, [date]);
  const year = date.slice(0, 4);
  const venues = new Set(items?.map(i => i.venue).filter(Boolean));
  const venue = commonest(items?.map(i => i.venue) ?? []); const place = commonest(items?.filter(i => !venue || i.venue === venue).map(i => i.location) ?? []);
  return <section className="date-view" aria-label={`Shows on ${date}`}>
    <nav className="crumbs eyebrow"><a href={`#/calendar/${year}`}>← {year}</a> · <a href="#/">The rack</a></nav>
    <div className="library-head"><h2>{date} <small>{items ? `${items.length} ${items.length === 1 ? 'tape' : 'tapes'}` : ''}</small></h2><p className="where">{venue}{place ? `, ${place}` : ''}{venues.size > 1 ? ` · ${venues.size} venues listed` : ''}</p></div>
    {error && <p className="status" role="alert">{error}</p>}
    {!items && !error && <p className="status" role="status">Pulling the tapes for this date…</p>}
    {items && items.length === 0 && <p className="shelf-empty">The collection has no tape for this date.</p>}
    {items && items.length > 0 && <div className="rack" role="list">{items.map(it => <div className="slot" role="listitem" key={it.id}><details className="spine-details">
      <summary className={`spine ${it.playable ? 'is-selected' : ''}`}><span className={kindTab(it.kind)}>{kindLabel(it.kind)}</span><span className={`label ${it.playable ? '' : it.library === 'imported' ? 'stock-aged' : 'stock-grey'}`}><span className="band">{it.venue || it.title || it.id}</span><span className="venue">{[it.taper && `taped by ${it.taper}`, it.transferer && `transfer ${it.transferer}`].filter(Boolean).join(' · ') || it.source || 'credits unrecorded'}</span>{it.playable ? <span className="sticker live">In the library</span> : it.library === 'imported' ? <span className="sticker warn">Imported, not playable</span> : <span className="sticker">Source only</span>}</span><span className="end">{it.duration ? clock(it.duration) : it.tracks.length ? `${it.tracks.length} tracks` : '—'}</span></summary>
      <div className="spine-body">
        <dl className="meta"><dt>Item</dt><dd>{it.id}</dd><dt>Source</dt><dd>{it.source || 'Not recorded'}</dd>{it.lineage && <><dt>Lineage</dt><dd>{it.lineage}</dd></>}{it.avg_rating > 0 && <><dt>Rating</dt><dd>{it.avg_rating.toFixed(1)} ({it.num_reviews} {it.num_reviews === 1 ? 'review' : 'reviews'})</dd></>}<dt>Downloads</dt><dd>{it.downloads.toLocaleString()}</dd></dl>
        {it.playable ? <button className="key primary" disabled={busy === it.id} onClick={() => onPlay(it.id)}>▶ {busy === it.id ? 'Loading…' : 'Play from the library'}</button> : <p className="status">Not in the library. Metadata only; nothing here authorizes playback or relay.</p>}
        {it.tracks.length > 0 && <ol className="tracks">{it.tracks.map(t => <li key={t.position}><span className="t">{t.title}</span><small>{clock(t.duration)}</small></li>)}</ol>}
        {it.warnings.length > 0 && <p className="status">{it.warnings.join(' · ')}</p>}
        <p><a href={`https://archive.org/details/${encodeURIComponent(it.id)}`} target="_blank" rel="noreferrer">Original item at Internet Archive ↗</a></p>
      </div>
    </details></div>)}</div>}
  </section>;
}
