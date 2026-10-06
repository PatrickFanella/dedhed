import { describe, expect, it } from "vitest";
import { radioCorrection, radioLabel } from "./radio";
import type { Playback } from "./types";
describe("private radio clock policy", () => {
  it("never seeks paused, unready or unbuffered media and corrects only beyond tolerance", () => {
    const playback = {
      mode: "radio",
      end: 20,
      receivedAt: 1000,
      latency: 0,
      station: { offset: 5 },
    } as Playback;
    const media = {
      paused: false,
      seeking: false,
      readyState: 4,
      currentTime: 0,
      buffered: { length: 1, start: () => 0, end: () => 20 },
    };
    expect(radioCorrection(media, playback, 2000)).toBe(6);
    expect(
      radioCorrection({ ...media, currentTime: 4 }, playback, 2000),
    ).toBeNull();
    expect(
      radioCorrection({ ...media, paused: true }, playback, 2000),
    ).toBeNull();
    expect(
      radioCorrection({ ...media, readyState: 2 }, playback, 2000),
    ).toBeNull();
    expect(
      radioCorrection(
        { ...media, buffered: { length: 0, start: () => 0, end: () => 0 } },
        playback,
        2000,
      ),
    ).toBeNull();
    expect(radioLabel("radio", "paused", "")).toBe(
      "Paused · Station continues",
    );
    expect(radioLabel("radio", "synchronized", "")).toBe(
      "On the station · Scheduled replay",
    );
  });
});
