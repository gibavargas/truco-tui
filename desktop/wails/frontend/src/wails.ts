import type { AppError, LocaleCode, RuntimeUpdate, SnapshotBundle } from "./types";

export const RUNTIME_UPDATE_EVENT = "truco:runtime:update";

type WailsBackend = {
  AcceptTruco(): Promise<AppError | null>;
  CloseSession(): Promise<AppError | null>;
  CreateHostSession(
    hostName: string,
    numPlayers: number,
    bindAddr: string,
    relayURL: string,
    transportMode: string,
  ): Promise<AppError | null>;
  JoinSession(key: string, playerName: string, desiredRole: string): Promise<AppError | null>;
  NewHand(): Promise<AppError | null>;
  PlayCard(cardIndex: number): Promise<AppError | null>;
  PlayFaceDownCard(cardIndex: number): Promise<AppError | null>;
  PollEvents(): Promise<Array<unknown>>;
  RefuseTruco(): Promise<AppError | null>;
  RequestReplacementInvite(targetSeat: number): Promise<AppError | null>;
  RequestTruco(): Promise<AppError | null>;
  Reset(): Promise<AppError | null>;
  RuntimeUpdateEventName(): Promise<string>;
  SendChat(text: string): Promise<AppError | null>;
  SetLocale(locale: LocaleCode): Promise<AppError | null>;
  Snapshot(): Promise<SnapshotBundle>;
  StartHostedMatch(): Promise<AppError | null>;
  StartOfflineGame(playerName: string, numPlayers: number): Promise<AppError | null>;
  Tick(maxSteps: number): Promise<AppError | null>;
  VoteHost(candidateSeat: number): Promise<AppError | null>;
};

type RuntimeAPI = {
  ClipboardSetText(text: string): Promise<void> | void;
  EventsOn(eventName: string, callback: (...data: unknown[]) => void): () => void;
};

declare global {
  interface Window {
    go?: {
      main?: {
        App?: WailsBackend;
      };
    };
    runtime?: RuntimeAPI;
  }
}

export async function snapshot(): Promise<SnapshotBundle> {
  return bridge().Snapshot();
}

export async function onRuntimeUpdate(callback: (update: RuntimeUpdate) => void): Promise<() => void> {
  const eventName = await bridge().RuntimeUpdateEventName();
  const runtime = window.runtime;
  if (!runtime) {
    throw new Error("Wails runtime unavailable");
  }
  return runtime.EventsOn(eventName || RUNTIME_UPDATE_EVENT, (...data) => {
    const update = data[0] as RuntimeUpdate | undefined;
    if (update?.bundle) {
      callback(update);
    }
  });
}

export async function invoke(action: string, payload: Record<string, unknown>): Promise<AppError | null> {
  const api = bridge();

  switch (action) {
    case "setLocale":
      return api.SetLocale(readLocale(payload.locale));
    case "startGame":
      return api.StartOfflineGame(stringValue(payload.name), numberValue(payload.numPlayers, 2));
    case "startOnlineHost":
      return api.CreateHostSession(
        stringValue(payload.name),
        numberValue(payload.numPlayers, 2),
        "",
        stringValue(payload.relay_url),
        transportValue(payload.transport_mode),
      );
    case "joinOnline":
      return api.JoinSession(
        stringValue(payload.key),
        stringValue(payload.name),
        stringValue(payload.role) || "auto",
      );
    case "startOnlineMatch":
      return api.StartHostedMatch();
    case "sendChat":
      return api.SendChat(stringValue(payload.message));
    case "sendHostVote":
      return api.VoteHost(numberValue(payload.slot, 0));
    case "requestReplacementInvite":
      return api.RequestReplacementInvite(numberValue(payload.slot, 0));
    case "play":
      if (payload.faceDown === true) {
        return api.PlayFaceDownCard(numberValue(payload.cardIndex, -1));
      }
      return api.PlayCard(numberValue(payload.cardIndex, -1));
    case "truco":
      return api.RequestTruco();
    case "accept":
      return api.AcceptTruco();
    case "refuse":
      return api.RefuseTruco();
    case "newHand":
      return api.NewHand();
    case "tick":
      return api.Tick(numberValue(payload.maxSteps, 12));
    case "closeSession":
      return api.CloseSession();
    case "reset":
      return api.Reset();
    default:
      throw new Error(`unsupported action: ${action}`);
  }
}

export async function copyText(value: string): Promise<void> {
  if (!value) {
    return;
  }
  if (window.runtime?.ClipboardSetText) {
    await window.runtime.ClipboardSetText(value);
    return;
  }
  if (navigator.clipboard) {
    await navigator.clipboard.writeText(value);
    return;
  }
  throw new Error("clipboard unavailable");
}

function bridge(): WailsBackend {
  const api = window.go?.main?.App;
  if (!api) {
    throw new Error("Wails bridge unavailable");
  }
  return api;
}

function readLocale(value: unknown): LocaleCode {
  return value === "en-US" ? "en-US" : "pt-BR";
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function numberValue(value: unknown, fallback: number): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && /^-?\d+$/.test(value)) {
    return Number.parseInt(value, 10);
  }
  return fallback;
}

function transportValue(value: unknown): string {
  switch (value) {
    case "tcp_tls":
      return "tcp_tls";
    case "relay_quic_v2":
      return "relay_quic_v2";
    default:
      return "";
  }
}
