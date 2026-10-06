import { audioPlaybackState, radioOffset } from "@subcult/archive-viewer/audio";
import type { Playback } from "./types";

export function radioState(
  media: HTMLAudioElement,
  playback: Playback,
  now: number,
) {
  const station =
    playback.mode === "radio" && playback.station
      ? {
          offset: playback.station.offset,
          receivedAt: playback.receivedAt!,
          latency: playback.latency!,
          duration: playback.end,
          programKey: `${playback.station.revision}:${playback.station.program.start}`,
        }
      : undefined;
  return audioPlaybackState(media, station, now);
}
export function radioCorrection(
  media: Pick<
    HTMLAudioElement,
    "paused" | "seeking" | "readyState" | "currentTime" | "buffered"
  >,
  playback: Playback,
  now: number,
) {
  if (
    playback.mode !== "radio" ||
    !playback.station ||
    media.paused ||
    media.seeking ||
    media.readyState < 3
  )
    return null;
  const target = radioOffset(
    playback.station.offset,
    playback.receivedAt!,
    playback.latency!,
    playback.end,
    now,
  );
  if (Math.abs(media.currentTime - target) <= 3) return null;
  for (let i = 0; i < media.buffered.length; i++)
    if (
      media.buffered.start(i) <= target &&
      media.buffered.end(i) >= Math.min(playback.end, target + 0.75)
    )
      return target;
  return null;
}
export function radioLabel(
  mode: Playback["mode"],
  state: string,
  label: string,
) {
  if (mode === "on-demand")
    return `On demand · ${label}${state === "playing" ? "" : ` · ${state}`}`;
  if (state === "synchronized") return "On the station · Scheduled replay";
  if (state === "paused") return "Paused · Station continues";
  return `Scheduled replay · ${state}`;
}
