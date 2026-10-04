import { useEffect, useRef, useState } from 'react';
import { useAudioControls } from '@subcult/archive-viewer/audio';
import { clock, currentTrack, setBounds, type Entry, type Playback, type Recording, type SetList, type Station } from './types';
import { CalendarView, DateView, OnThisDay, useRoute } from './shows';

const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI'];
const SIDES = 'ABCDEF';

async function get<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) { const data = await res.json().catch(() => null); throw new Error(data?.error ?? "Dedhed isn't answering. Try again in a moment."); }
  return res.json() as Promise<T>;
}

/** The brand tab at the end of a spine names the source type; synthetic tapes get the chrome tab. */
function tabClass(r: Recording): string {
  if (r.synthetic) return 'tab chrome';
  return r.kind.toLowerCase() === 'sbd' ? 'tab gold' : 'tab';
}
function setNumerals(r: Recording): string { return r.sets.filter(s => s.reviewed).map((_, i) => ROMAN[i] ?? String(i + 1)).join(' · '); }
function setsForDisplay(r: Recording): SetList[] {
  return r.sets.length ? r.sets : [{ title: 'Complete recording · set boundaries unreviewed', first: 0, last: r.tracks.length - 1, reviewed: false }];
}

export default function App() {
  const [items, setItems] = useState<Entry[]>([]), [total, setTotal] = useState(0);
  const [query, setQuery] = useState(''), [year, setYear] = useState(''), [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Entry | null>(null), [station, setStation] = useState<Station | null>(null);
  const [loading, setLoading] = useState(true), [catalogError, setCatalogError] = useState(''), [stationError, setStationError] = useState('');
  const [playback, setPlayback] = useState<Playback | null>(null);
  const route = useRoute(); const [opening, setOpening] = useState('');
  const audio = useRef<HTMLAudioElement>(null), pendingSeek = useRef<number | null>(null), operation = useRef(0), advancing = useRef(false);
  const controls=useAudioControls({audio,start:playback?.start??0,end:playback?.end??1,initialVolume:0.7,
    onEnded:()=>{if(playback?.mode==='radio'&&!advancing.current){advancing.current=true;void tune();}}});
  const {position,playing,pending,error:playerError,volume,seek,toggle,setError:setPlayerError,setPlaying,setPending}=controls;


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
    const token = operation.current; advancing.current = false; stopAudio(); pendingSeek.current = target; setPlayback(next); controls.reset(target);
    a.src = `/api/audio/${encodeURIComponent(next.entry.recording.id)}`; a.volume = volume; a.load();
    try { await a.play(); } catch (e) { if (token === operation.current && e instanceof Error && e.name !== 'AbortError') setPlayerError("The tape didn't start. Press play to try again."); }
  }
  async function playSet(entry: Entry, set?: SetList, trackStart?: number) {
    operation.current++; const [start, end] = setBounds(entry.recording, set);
    await load({ entry, mode: 'on-demand', start, end, label: set?.title ?? 'Full recording' }, trackStart ?? start);
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
    } catch (e) { if (op === operation.current) { setPlayerError(e instanceof Error ? e.message : "The station isn't answering. Try again, or pull a tape from the rack."); setPending(false); } }
  }
  function metadataReady() {
    const a = audio.current; if (!a) return;
    if (pendingSeek.current !== null) { seek(pendingSeek.current); pendingSeek.current = null; setPending(false); }
  }
  function timeUpdate() { if(pendingSeek.current===null&&playback)controls.events.onTimeUpdate(); }

  async function openFromIndex(id: string) {
    setOpening(id); setPlayerError('');
    try { const entry = await get<Entry>(`/api/recordings/${encodeURIComponent(id)}`); setSelected(entry); await playSet(entry); }
    catch (e) { setPlayerError(e instanceof Error ? e.message : "That tape isn't in the library."); }
    finally { setOpening(''); }
  }
  const r = selected?.recording;
  const track = playback ? currentTrack(playback.entry.recording, position) : null;
  const onAir = station ? items.find(i => i.recording.id === station.program.recording_id)?.recording : null;
  const nextName = station?.next ? items.find(i => i.recording.id === station.next?.recording_id)?.recording.venue : null;
  const playingHere = playback && r && playback.entry.recording.id === r.id;

  return <>
    <header className="masthead">
      <a href="/" className="wordmark" aria-label="Dedhed home"><i aria-hidden="true" />dedhed</a>
      <nav className="masthead-nav" aria-label="Sections"><a href="#/" aria-current={route.view === 'home' ? 'page' : undefined}>The rack</a><a href="#/calendar" aria-current={route.view !== 'home' ? 'page' : undefined}>Calendar</a></nav>
      {station ? <span className="sticker live">On air{onAir ? ` · ${onAir.date}` : ''}</span> : <span className="sticker">Off air</span>}
    </header>
    <main>
      {route.view === 'calendar' && <CalendarView year={route.year} />}
      {route.view === 'date' && <DateView date={route.date} onPlay={id => { void openFromIndex(id); }} busy={opening} />}
      {route.view === 'home' && <>
      <section className="intro">
        <div>
          <p className="eyebrow">Grateful Dead tapes · {import.meta.env.VITE_DEDHED_AUDIENCE === 'public' ? 'free listening' : 'private pilot'}</p>
          <h1>Find your way<br />into a show.</h1>
          <p className="lede">Pull a tape off the shelf and play one set, or let the whole show run. Or tune in to the station and hear whatever everyone else is hearing.</p>
        </div>
        <aside className={`station ${playback?.mode === 'radio' && playing ? 'is-playing' : ''}`} aria-label="Shared radio">
          <div className="head"><span className="sticker live">On air</span><span className="reel" aria-hidden="true" style={{ width: 28, height: 28, borderWidth: 2 }} /></div>
          <p className="eyebrow">Shared radio</p>
          <h2>Already in<br />progress.</h2>
          <p>{station ? (onAir ? `${onAir.venue || onAir.title} · ${onAir.date}` : 'Tune in and you land where the reels are.') : stationError || 'Checking the station…'}</p>
          <button className="key primary" onClick={() => { void tune(); }} disabled={!station}>Tune in ↗</button>
          {station?.next && <small>UP NEXT · {nextName ?? 'Next tape on the schedule'}</small>}
        </aside>
      </section>
      <OnThisDay />

      <section className="library" aria-label="Concert library">
        <div className="library-head">
          <h2>The rack <small>{total.toString().padStart(2, '0')} tapes</small></h2>
          <div className="filters">
            <label className="field"><span>Find a show</span><input type="search" placeholder="Date, venue, city…" value={query} onChange={e => { setQuery(e.target.value); setPage(1); }} /></label>
            <label className="field"><span>Year</span><input inputMode="numeric" placeholder="All" maxLength={4} value={year} onChange={e => { setYear(e.target.value); setPage(1); }} /></label>
          </div>
        </div>
        <div className="columns">
          <div>
            {catalogError && <p className="status" role="alert">{catalogError}</p>}
            {loading && <p className="status" role="status">Pulling tapes off the shelf…</p>}
            {!loading && !catalogError && items.length === 0 && <p className="shelf-empty">Nothing on the shelf matches. Try another date or venue.</p>}
            {items.length > 0 && <div className="rack" role="list">
              {items.map(entry => { const rec = entry.recording; const isSel = selected?.recording.id === rec.id; const unreviewed = !rec.sets.some(s => s.reviewed); return <div className="slot" role="listitem" key={rec.id}>
                <button className="spine" onClick={() => setSelected(entry)} aria-pressed={isSel}>
                  <span className={tabClass(rec)}>{rec.synthetic ? 'TEST' : rec.kind.toUpperCase()}</span>
                  <span className={`label ${rec.synthetic ? 'stock-grey' : unreviewed ? 'stock-aged' : ''}`}>
                    <span className="band">{rec.date}</span>
                    {setNumerals(rec) && <span className="set">{setNumerals(rec)}</span>}
                    <span className="venue">{rec.venue || 'Venue unrecorded'}{rec.location ? `, ${rec.location}` : ''}</span>
                    {rec.synthetic ? <span className="sticker red">Synthetic</span> : unreviewed ? <span className="sticker warn">Sets unreviewed</span> : null}
                  </span>
                  <span className="end">{clock(rec.duration)}</span>
                </button>
              </div>; })}
              {items.length < 6 && Array.from({ length: 6 - items.length }, (_, i) => <div className="slot empty" aria-hidden="true" key={`empty-${i}`} />)}
            </div>}
            {total > 20 && <nav className="pager" aria-label="Catalog pages"><button className="key ghost" disabled={page === 1} onClick={() => setPage(page - 1)}>← Previous</button><span>Shelf {page} of {Math.ceil(total / 20)}</span><button className="key ghost" disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>Next →</button></nav>}
          </div>

          <article className="detail" aria-label="Recording details">{r && selected ? <>
            <div className={`cassette ${playingHere && playing ? 'is-playing' : ''}`}>
              <div className="face">
                <p className="eyebrow">{r.synthetic ? 'Synthetic development recording' : `${r.kind} recording`}</p>
                <span className="side">A</span>
                <h2>{r.venue || r.title}</h2>
                <div className="stamp">{r.date}<b>{clock(r.duration)}</b></div>
                <div className="window" aria-hidden="true"><span className="reel" /><span className="tape" /><span className="reel" /></div>
                <p className="line">{r.location || 'Location unrecorded'} · <b>Taped by</b> {r.taper || 'not recorded'} · <b>Transferred by</b> {r.transferer || 'not recorded'} · {r.sets.some(s => s.reviewed) ? 'Set boundaries reviewed' : 'Set boundaries unreviewed'}</p>
              </div>
              <div className="controls">
                <button className="key primary" onClick={() => { void playSet(selected); }}>▶ Full recording · {clock(r.duration)}</button>
                {r.sets.filter(s => s.reviewed).map((set, i) => <button className="key ghost" key={set.title} onClick={() => { void playSet(selected, set); }}>▶ {ROMAN[i] ?? set.title}</button>)}
              </div>
            </div>
            <section className="jcard" aria-label="Sets and tracks">
              {setsForDisplay(r).map((set, i) => <div className="side-panel" key={i}>
                <div className="side-head"><span className="side">{SIDES[i] ?? '·'}</span><h3>{set.title}</h3><span className="eyebrow">{clock(r.tracks.slice(set.first, set.last + 1).reduce((n, t) => n + t.duration, 0))}</span>{set.reviewed && <button className="key" onClick={() => { void playSet(selected, set); }}>Play</button>}</div>
                <ol className="tracks" start={set.first + 1}>{r.tracks.slice(set.first, set.last + 1).map(t => <li key={t.position} className={playingHere && track?.position === t.position ? 'now' : ''}><button onClick={() => { void playSet(selected, set.reviewed ? set : undefined, t.start); }} aria-current={playingHere && track?.position === t.position ? 'true' : undefined}><span className="t">{t.title}</span><small>{clock(t.duration)}</small></button></li>)}</ol>
              </div>)}
              <details className="credits">
                <summary>Recording credits &amp; source</summary>
                <dl><div><dt>Taped by</dt><dd>{r.taper || 'Not recorded'}</dd></div><div><dt>Transferred by</dt><dd>{r.transferer || 'Not recorded'}</dd></div><div><dt>Source lineage</dt><dd>{r.source}</dd></div></dl>
                {!r.synthetic && <p><a href={`https://archive.org/details/${encodeURIComponent(r.id)}`} target="_blank" rel="noreferrer">Original recording at Internet Archive ↗</a></p>}
                {selected.review.notes && <p>{selected.review.notes}</p>}
                {r.warnings.map(w => <div className="note" key={w}><span className="sticker warn">Note</span><span>{w}</span></div>)}
              </details>
            </section>
          </> : <p className="shelf-empty">Pull a tape from the rack to see its sets and credits.</p>}</article>
        </div>
      </section>
      </>}
      <footer className="footer">Labelled in pencil where we weren&apos;t sure.<span>Source credits stay with every tape.</span></footer>
      {import.meta.env.VITE_DEDHED_AUDIENCE === 'public' && <section aria-label="Trading policy" className="intro">
        <div><h2>Keep the tapes free.</h2><p>This is a free, noncommercial audience-recording library. No advertising, paid access, sale of listener data, or commercial promotion funds this site. Copyright belongs to the performers, writers and publishers. We respect their rights and remove access when permission is withdrawn.</p>
        <p>Taper and transfer credits stay with each recording. The band can withdraw permission for noncommercial digital exchange. Read the <a href="https://archive.org/details/GratefulDead">Grateful Dead trading notice</a> and <a href="https://archivesupport.zendesk.com/hc/en-us/articles/360014759692-Rights">Internet Archive policy</a>. For a correction or removal, contact <a href="mailto:info@subcult.tv">info@subcult.tv</a>.</p></div>
      </section>}
    </main>

    <section className={`deck ${playing ? 'is-playing' : ''}`} aria-label="Audio player">
      <audio ref={audio} preload="metadata" onLoadedMetadata={metadataReady} onTimeUpdate={timeUpdate} onPlaying={() => { setPlaying(true); setPending(false); setPlayerError(''); }} onPause={() => setPlaying(false)} onWaiting={() => setPending(true)} onError={() => { setPending(false); setPlaying(false); setPlayerError("The audio source isn't answering. Try again, or pull another tape."); }} onEnded={controls.events.onEnded} />
      {playerError && <p className="player-error" role="alert">{playerError} {playback && <button onClick={() => { if (playback.mode === 'radio') void tune(); else void load(playback, Math.max(playback.start, Math.min(position, playback.end - 0.1))); }}>Try again</button>}</p>}
      <div className="inner">
        <div className="window" aria-hidden="true"><span className="reel" /><span className="tape" /><span className="reel" /></div>
        <div className="keys" role="group" aria-label="Transport">
          <button className="key" onClick={() => { void toggle(); }} disabled={!playback} aria-pressed={playing} aria-label={playing ? 'Pause' : 'Play'}>{playing ? 'Ⅱ' : '▶'}</button>
        </div>
        <div className="now-playing">
          <p className="eyebrow">{playback ? (playback.mode === 'radio' ? playback.label : `On demand · ${playback.label}`) : 'Nothing in the deck'}</p>
          <strong>{track?.title ?? 'Pull a tape or tune in'}</strong>
          <small>{playback?.entry.recording.venue ?? 'Grateful Dead tapes'}{playback ? ` · ${playback.entry.recording.date}` : ''}{pending && playback ? ' · Buffering…' : ''}</small>
        </div>
        <div className="transport">
          <span>{clock(position - (playback?.start ?? 0))}</span>
          <input aria-label="Playback position" type="range" min={playback?.start ?? 0} max={playback?.end ?? 1} step="0.1" value={position} disabled={!playback} onChange={e => seek(Number(e.target.value))} />
          <span>{clock(playback ? playback.end - playback.start : 0)}</span>
          {playback?.mode === 'radio' && <button className="return" onClick={() => { void tune(); }}>Return to live</button>}
        </div>
        <label className="volume">Volume<input type="range" min="0" max="1" step="0.05" value={volume} onChange={e => { controls.changeVolume(Number(e.target.value)); }} /></label>
      </div>
    </section>
  </>;
}
