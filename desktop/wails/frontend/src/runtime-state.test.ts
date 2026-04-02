import test from "node:test";
import assert from "node:assert/strict";

import {
  expectedModesForAction,
  recoveryStateForBundle,
  shouldApplyIncomingSequence,
  viewForMode,
} from "./runtime-state";

test("viewForMode maps runtime modes explicitly", () => {
  assert.equal(viewForMode("idle"), "setup");
  assert.equal(viewForMode("host_lobby"), "lobby");
  assert.equal(viewForMode("client_lobby"), "lobby");
  assert.equal(viewForMode("offline_match"), "game");
  assert.equal(viewForMode("host_match"), "game");
  assert.equal(viewForMode("client_match"), "game");
});

test("expectedModesForAction encodes transition-critical flows", () => {
  assert.deepEqual(expectedModesForAction("startGame"), ["offline_match"]);
  assert.deepEqual(expectedModesForAction("startOnlineHost"), ["host_lobby"]);
  assert.deepEqual(expectedModesForAction("joinOnline"), ["client_lobby", "client_match"]);
  assert.deepEqual(expectedModesForAction("reset"), ["idle"]);
});

test("shouldApplyIncomingSequence rejects stale events but accepts snapshots", () => {
  assert.equal(shouldApplyIncomingSequence(7, 6, "event"), false);
  assert.equal(shouldApplyIncomingSequence(7, 7, "event"), true);
  assert.equal(shouldApplyIncomingSequence(7, 3, "snapshot"), true);
});

test("recoveryStateForBundle flags missing lobby and match snapshots", () => {
  assert.equal(recoveryStateForBundle({ mode: "host_lobby", lobby: null }), "waiting_lobby");
  assert.equal(recoveryStateForBundle({ mode: "offline_match", match: null }), "waiting_match");
  assert.equal(
    recoveryStateForBundle({
      mode: "offline_match",
      match: { CurrentHand: { Round: 1 }, Players: [{ ID: 0 }, { ID: 1 }] },
    }),
    null,
  );
});
