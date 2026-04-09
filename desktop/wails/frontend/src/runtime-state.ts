export type ViewName = "setup" | "lobby" | "game";
export type BundleLike = {
  mode?: string;
  lobby?: unknown;
  match?: {
    CurrentHand?: unknown;
    Players?: unknown[];
  } | null;
};

export type RecoveryState = "waiting_lobby" | "waiting_match" | null;

export function viewForMode(mode: string): ViewName {
  switch (mode) {
    case "host_lobby":
    case "client_lobby":
      return "lobby";
    case "offline_match":
    case "host_match":
    case "client_match":
      return "game";
    default:
      return "setup";
  }
}

export function expectedModesForAction(action: string): string[] {
  switch (action) {
    case "startGame":
      return ["offline_match"];
    case "startOnlineHost":
      return ["host_lobby"];
    case "joinOnline":
      return ["client_lobby", "client_match"];
    case "startOnlineMatch":
      return ["host_match"];
    case "closeSession":
    case "reset":
      return ["idle"];
    default:
      return [];
  }
}

export function shouldApplyIncomingSequence(
  currentSequence: number,
  incomingSequence: number,
  source: "snapshot" | "event",
): boolean {
  if (source === "snapshot") {
    return true;
  }
  return incomingSequence >= currentSequence;
}

export function recoveryStateForBundle(bundle: BundleLike | null): RecoveryState {
  const view = viewForMode(bundle?.mode || "idle");
  if (view === "lobby" && !bundle?.lobby) {
    return "waiting_lobby";
  }
  if (view !== "game") {
    return null;
  }
  const match = bundle?.match;
  if (!match || !match.CurrentHand || !Array.isArray(match.Players) || match.Players.length === 0) {
    return "waiting_match";
  }
  return null;
}
