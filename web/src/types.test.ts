import { describe, expect, it } from 'vitest';
import { clock, currentTrack, monthGrid, setBounds, type Recording } from './types';

const r = { duration: 40, tracks: [0, 1, 2, 3].map(i => ({ title: String(i), position: i, start: i * 10, duration: 10 })) } as Recording;
describe('playback boundaries', () => {
  it('uses reviewed set tracks, including exact ends', () => {
    expect(setBounds(r, { title: 'II', first: 2, last: 3, reviewed: true })).toEqual([20, 40]);
    expect(setBounds(r)).toEqual([0, 40]);
    expect(() => setBounds(r, { title: 'disc', first: 0, last: 1, reviewed: false })).toThrow();
    expect(currentTrack(r, 10).position).toBe(1);
  });
  it('formats concert and set durations', () => {
    expect(clock(3601)).toBe('1:00:01'); expect(clock(65)).toBe('1:05');
  });
});

it('month grid starts on Monday and pads to whole weeks', () => {
  const may77 = monthGrid(1977, 5);
  expect(may77.length % 7).toBe(0);
  expect(may77.slice(0, 7)).toEqual(['', '', '', '', '', '', '1977-05-01']); // 1 May 1977 was a Sunday
  expect(may77.filter(Boolean).length).toBe(31);
  expect(monthGrid(2024, 2).filter(Boolean).length).toBe(29);
});

import { commonest, parseRoute } from './shows';
it('routes and the commonest venue', () => {
  expect(parseRoute('#/calendar/1977')).toEqual({ view: 'calendar', year: 1977 });
  expect(parseRoute('#/calendar')).toEqual({ view: 'calendar', year: null });
  expect(parseRoute('#/date/1977-05-08')).toEqual({ view: 'date', date: '1977-05-08' });
  expect(parseRoute('#/date/1977-5-8')).toEqual({ view: 'home' });
  expect(commonest(['Barton Hall (synthetic)', 'Barton Hall', '', 'Barton Hall'])).toBe('Barton Hall');
  expect(commonest([])).toBe('');
});
