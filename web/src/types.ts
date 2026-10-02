export type Track = { title: string; position: number; start: number; duration: number };
export type SetList = { title: string; first: number; last: number; reviewed: boolean };
export type Recording = {
  id: string; concert_id: string; artist: string; date: string; venue: string;
  location: string; title: string; kind: string; source: string; taper: string;
  transferer: string; tracks: Track[]; sets: SetList[]; duration: number;
  warnings: string[]; synthetic: boolean;
};
export type Entry = { recording: Recording; review: { notes: string }; media: string };
export type Program = { recording_id: string; start: string; duration: number };
export type Station = { server_time: string; revision: string; program: Program; offset: number; track: number; next?: Program };
export type Playback = { entry: Entry; mode: 'radio' | 'on-demand'; start: number; end: number; label: string; station?: Station };

export function clock(seconds: number): string {
  const n = Math.max(0, Math.floor(seconds));
  return n >= 3600 ? `${Math.floor(n / 3600)}:${String(Math.floor(n / 60) % 60).padStart(2, '0')}:${String(n % 60).padStart(2, '0')}` : `${Math.floor(n / 60)}:${String(n % 60).padStart(2, '0')}`;
}
export function setBounds(r: Recording, set?: SetList): [number, number] {
  if (!set) return [0, r.duration];
  if (!set.reviewed || set.first < 0 || set.last >= r.tracks.length || set.first > set.last) throw new Error('Set boundaries are unavailable');
  return [r.tracks[set.first].start, r.tracks[set.last].start + r.tracks[set.last].duration];
}
export function currentTrack(r: Recording, position: number): Track {
  return r.tracks.find(t => position >= t.start && position < t.start + t.duration) ?? r.tracks[r.tracks.length - 1];
}

export type SourceItem = {
  id: string; show_date: string; date_text: string; artist: string; venue: string; location: string; title: string;
  kind: string; source: string; lineage: string; taper: string; transferer: string; downloads: number; avg_rating: number;
  num_reviews: number; item_size: number; tracks: Track[]; songs: string; duration: number; warnings: string[]; snapshot_sha256: string;
};
export type DateItem = SourceItem & { playable: boolean; library: '' | 'imported' | 'playable' };
export type YearCount = { year: number; items: number; dates: number };
export type DayCount = { date: string; items: number; venue: string };
export type SearchHit = { date: string; venue: string; items: SourceItem[] };

export const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

/** Calendar cells for a month, Monday first: '' for padding, else YYYY-MM-DD. */
export function monthGrid(year: number, month: number): string[] {
  const first = new Date(Date.UTC(year, month - 1, 1));
  const lead = (first.getUTCDay() + 6) % 7;
  const days = new Date(Date.UTC(year, month, 0)).getUTCDate();
  const cells: string[] = Array.from({ length: lead }, () => '');
  for (let d = 1; d <= days; d++) cells.push(`${year}-${String(month).padStart(2, '0')}-${String(d).padStart(2, '0')}`);
  while (cells.length % 7) cells.push('');
  return cells;
}
