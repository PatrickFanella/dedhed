import { describe, expect, it } from 'vitest';
import { clock, currentTrack, setBounds, type Recording } from './types';

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
