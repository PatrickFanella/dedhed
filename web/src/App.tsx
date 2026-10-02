import { useEffect, useRef, useState } from 'react';
import { clock, currentTrack, setBounds, type Entry, type Playback, type SetList, type Station } from './types';

async function get<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) { const data = await res.json().catch(() => null); throw new Error(data?.error ?? 'Unable to load this recording'); }
  return res.json() as Promise<T>;
}

export default function App() {
  const [items, setItems] = useState<Entry[]>([]), [total, setTotal] = useState(0);
  const [query, setQuery] = useState(''), [year, setYear] = useState(''), [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Entry | null>(null), [station, setStation] = useState<Station | null>(null);
  const [loading, setLoading] = useState(true), [catalogError, setCatalogError] = useState(''), [stationError, setStationError] = useState('');
  const [playback, setPlayback] = useState<Playback | null>(null), [playing, setPlaying] = useState(false);
  const [position, setPosition] = useState(0), [playerError, setPlayerError] = useState(''), [volume, setVolume] = useState(0.7), [pending, setPending] = useState(false);
  const audio = useRef<HTMLAudioElement>(null), pendingSeek = useRef<number | null>(null), operation = useRef(0), advancing = useRef(false);

  useEffect(() => {
    let active = true;
    setLoading(true); setCatalogError('');
    const params = new URLSearchParams({ q: query, year, page: String(page) });
    get<{ items: Entry[]; total: number }>(`/api/recordings?${params}`).then(data => {
      if (active) { setItems(data.items); setTotal(data.total); setSelected(prev => prev ?? data.items[0] ?? null); }
    }).catch(e => { if (active) setCatalogError(String(e.message)); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [query, year, page]);

  useEffect(() => {
    let active = true;
    const poll = () => get<Station>('/api/station').then(data => { if (active) { setStation(data); setStationError(''); } }).catch(e => { if (active) { setStation(null); setStationError(e.message); } });
    void poll(); const timer = window.setInterval(() => { void poll(); }, 5000);
    return () => { active = false; window.clearInterval(timer); };
  }, []);

  function stopAudio() { audio.current?.pause(); setPlaying(false); setPlayerError(''); }
  async function load(next: Playback, target: number) {
    const a = audio.current; if (!a) return;
    const token = operation.current; advancing.current = false; stopAudio(); pendingSeek.current = target; setPlayback(next); setPosition(target); setPending(true);
    a.src = `/api/audio/${encodeURIComponent(next.entry.recording.id)}`; a.volume = volume; a.load();
    try { await a.play(); } catch (e) { if (token === operation.current && e instanceof Error && e.name !== 'AbortError') setPlayerError('Playback could not start. Press Play to retry.'); }
  }
  async function playSet(entry: Entry, set?: SetList, trackStart?: number) {
    operation.current++; const [start, end] = setBounds(entry.recording, set);
    await load({ entry, mode: 'on-demand', start, end, label: set?.title ?? 'Complete show' }, trackStart ?? start);
  }
  async function tune() {
    const op = ++operation.current; setPlayerError(''); setPending(true);
    try {
      const began = performance.now(); const live = await get<Station>('/api/station');
      const entry = await get<Entry>(`/api/recordings/${encodeURIComponent(live.program.recording_id)}`);
      if (op !== operation.current) return;
      const target = Math.min(live.program.duration - 0.1, live.offset + (performance.now() - began) / 1000);
      setSelected(entry); setStation(live);
      await load({ entry, mode: 'radio', start: 0, end: live.program.duration, label: 'Shared radio', station: live }, target);
    } catch (e) { if (op === operation.current) { setPlayerError(e instanceof Error ? e.message : 'Station unavailable'); setPending(false); } }
  }
  async function toggle() {
    const a = audio.current; if (!a || !playback) return;
    if (!a.paused) { a.pause(); return; }
    if (position >= playback.end - 0.1) a.currentTime = playback.start;
    try { await a.play(); } catch { setPlayerError('Audio unavailable. Retry playback or choose another recording.'); }
  }
  function metadataReady() {
    const a = audio.current; if (!a) return;
    if (pendingSeek.current !== null) { a.currentTime = pendingSeek.current; pendingSeek.current = null; }
  }
  function timeUpdate() {
    const a = audio.current; if (!a || !playback || pendingSeek.current !== null) return;
    setPosition(a.currentTime);
    if (a.currentTime >= playback.end - 0.04) {
      a.pause();
      if (playback.mode === 'radio' && !advancing.current) { advancing.current = true; void tune(); }
      else { a.currentTime = playback.end; setPosition(playback.end); }
    }
  }
  function seek(value: number) { if (audio.current && playback) { audio.current.currentTime = Math.min(playback.end, Math.max(playback.start, value)); setPosition(audio.current.currentTime); } }
  const r = selected?.recording;
  const track = playback ? currentTrack(playback.entry.recording, position) : null;
  const nextName = station?.next ? items.find(i => i.recording.id === station.next?.recording_id)?.recording.venue : null;

  return <>
    <header className="masthead"><a href="/" className="wordmark" aria-label="Dedhed home">dedhed<span>●</span></a><p>LIVE TAPES / SHARED AIR</p><span className="edition">PRIVATE PILOT · VOL. 01</span></header>
    <main>
      <section className="intro"><div><p className="eyebrow">THE GRATEFUL DEAD ARCHIVE</p><h1>Find your way<br />into a show.</h1><p className="lede">Choose a night. Play a set. Or tune in and see where the music takes us.</p></div>
        <aside className="station"><p className="eyebrow"><span className="live-dot" /> SHARED RADIO</p><h2>One station.<br />The same moment.</h2><p>{station ? 'Join the concert already in progress.' : stationError || 'Checking the station…'}</p><button className="primary" onClick={() => { void tune(); }} disabled={!station}>Tune in <span aria-hidden="true">↗</span></button>{station?.next && <small>UP NEXT · {nextName ?? 'Next scheduled concert'}</small>}</aside>
      </section>

      <section className="library" aria-label="Concert library">
        <div className="library-head"><h2>The tape shelf <span>{total.toString().padStart(2, '0')}</span></h2><p>CONCERTS, SETS & SOURCE NOTES</p></div>
        <div className="columns"><div className="shelf">
          <div className="filters"><label>Find a show<input type="search" placeholder="Date, venue, city…" value={query} onChange={e => { setQuery(e.target.value); setPage(1); }} /></label><label>Year<input inputMode="numeric" placeholder="All years" maxLength={4} value={year} onChange={e => { setYear(e.target.value); setPage(1); }} /></label></div>
          {catalogError && <p role="alert">{catalogError}</p>}{loading && <p role="status">Loading the tape shelf…</p>}
          {!loading && !catalogError && items.length === 0 && <p className="empty">No shows match this search. Try another date or venue.</p>}
          <div className="recordings">{items.map((entry, i) => <button key={entry.recording.id} className={`tape ${selected?.recording.id === entry.recording.id ? 'selected' : ''}`} onClick={() => setSelected(entry)} aria-pressed={selected?.recording.id === entry.recording.id}>
            <span className="tape-number">{String((page - 1) * 20 + i + 1).padStart(2, '0')}</span><span className="tape-title"><strong>{entry.recording.date}</strong><span>{entry.recording.venue || 'Venue unrecorded'}</span><small>{entry.recording.location || 'Location unrecorded'}</small>{entry.recording.synthetic && <em>TEST TONES · SYNTHETIC</em>}</span><span className="tape-length">{clock(entry.recording.duration)}<span aria-hidden="true">↗</span></span>
          </button>)}</div>
          {total > 20 && <nav aria-label="Catalog pages"><button disabled={page === 1} onClick={() => setPage(page - 1)}>Previous</button><span>Page {page}</span><button disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>Next</button></nav>}
        </div>
        <article className="detail" aria-label="Recording details">{r && selected ? <>
          <p className="eyebrow">{r.synthetic ? 'SYNTHETIC DEVELOPMENT RECORDING' : `${r.kind.toUpperCase()} RECORDING`}</p><h2>{r.venue || r.title}</h2><p className="show-date">{r.date} <span> / </span> {r.location}</p>
          <button className="outline" onClick={() => { void playSet(selected); }}>Play complete show <span>{clock(r.duration)}</span></button>
          {(r.sets.length ? r.sets : [{ title: 'Complete recording · set boundaries unreviewed', first: 0, last: r.tracks.length - 1, reviewed: false }]).map((set, i) => <section className="set" key={i}><div className="set-head"><h3>{set.title}</h3>{set.reviewed && <button onClick={() => { void playSet(selected, set); }}>Play set <span aria-hidden="true">▶</span></button>}</div><ol start={set.first + 1}>{r.tracks.slice(set.first, set.last + 1).map(t => <li key={t.position}><button onClick={() => { void playSet(selected, set.reviewed ? set : undefined, t.start); }}><span>{t.title}</span><small>{clock(t.duration)}</small></button></li>)}</ol></section>)}
          <details className="credits"><summary>Recording credits & source</summary><dl><dt>Taped by</dt><dd>{r.taper || 'Not recorded'}</dd><dt>Transferred by</dt><dd>{r.transferer || 'Not recorded'}</dd><dt>Source lineage</dt><dd>{r.source}</dd></dl>{!r.synthetic && <a href={`https://archive.org/details/${encodeURIComponent(r.id)}`} target="_blank" rel="noreferrer">Original recording at Internet Archive ↗</a>}<p>{selected.review.notes}</p>{r.warnings.map(w => <p key={w}>{w}</p>)}</details>
        </> : <p className="empty">Choose a concert from the shelf to explore its sets and recordings.</p>}</article></div>
      </section>
      <footer className="footer">For the nights that keep going.<span>Source credits stay with every tape.</span></footer>
    </main>
    <section className="player" aria-label="Audio player">
      <audio ref={audio} preload="metadata" onLoadedMetadata={metadataReady} onTimeUpdate={timeUpdate} onPlaying={() => { setPlaying(true); setPending(false); setPlayerError(''); }} onPause={() => setPlaying(false)} onWaiting={() => setPending(true)} onError={() => { setPending(false); setPlaying(false); setPlayerError('Audio source unavailable. Retry or choose another recording.'); }} onEnded={() => { if (playback?.mode === 'radio' && !advancing.current) { advancing.current = true; void tune(); } else setPlaying(false); }} />
      <button className="play-toggle" onClick={() => { void toggle(); }} disabled={!playback} aria-label={playing ? 'Pause playback' : 'Play playback'}>{playing ? 'Ⅱ' : '▶'}</button>
      <div className="playing-info"><span className="eyebrow">{playback ? `${playback.mode === 'radio' ? 'RADIO' : 'ON DEMAND'} / ${playback.label}` : 'READY WHEN YOU ARE'}</span><strong>{track?.title ?? 'Choose a set or tune in'}</strong><small>{playback?.entry.recording.venue ?? 'The Grateful Dead archive'}{pending && playback ? ' · Buffering…' : ''}</small></div>
      <div className="transport"><div><span>{clock(position - (playback?.start ?? 0))}</span><input aria-label="Playback position" type="range" min={playback?.start ?? 0} max={playback?.end ?? 1} step="0.1" value={position} disabled={!playback} onChange={e => seek(Number(e.target.value))} /><span>{clock(playback ? playback.end - playback.start : 0)}</span></div>{playback?.mode === 'radio' && <button onClick={() => { void tune(); }}>Return to live</button>}</div>
      <label className="volume">Volume<input type="range" min="0" max="1" step="0.05" value={volume} onChange={e => { const n = Number(e.target.value); setVolume(n); if (audio.current) audio.current.volume = n; }} /></label>
      {playerError && <p className="player-error" role="alert">{playerError} {playback && <button onClick={() => { if (playback.mode === 'radio') void tune(); else void load(playback, Math.max(playback.start, Math.min(position, playback.end - 0.1))); }}>Retry playback</button>}</p>}
    </section>
  </>;
}
