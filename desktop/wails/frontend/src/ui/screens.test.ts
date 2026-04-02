import test from "node:test";
import assert from "node:assert/strict";

import { renderGameScreen } from "./game-screen";
import { renderLobbyScreen } from "./lobby-screen";
import { renderSetupScreen } from "./setup-screen";
import type { Card, MatchSnapshot, RuntimeEvent, SnapshotBundle } from "../types";

const escapeHtml = (value: string): string => value;
const t = (key: string, ...args: Array<string | number>): string => [key, ...args].join(" ");
const busyAttr = (): string => "";
const buttonLabel = (_formId: string, label: string): string => label;
const renderMetric = (label: string, value: string): string => `<metric>${label}:${value}</metric>`;
const renderEventFeed = (): string => "event-feed";
const renderCard = (card: Card, size: "tiny" | "small" | "regular" = "regular"): string =>
  `<card size="${size}">${card.Rank}-${card.Suit}</card>`;
const protocolLabel = (): string => "v2";
const cardLabel = (card: Card): string => `${card.Rank} of ${card.Suit}`;
const playerName = (match: MatchSnapshot, playerId: number): string =>
  match.Players.find((player) => player.ID === playerId)?.Name || "?";
const teamScore = (match: MatchSnapshot, team: number): number => match.MatchPoints[String(team)] || 0;
const localTeam = (match: MatchSnapshot, bundle: SnapshotBundle): number =>
  match.Players.find((player) => player.ID === bundle.ui.actions.local_player_id)?.Team || 0;
const nextStake = (current: number): number => ({ 1: 3, 3: 6, 6: 9, 9: 12 }[current] || current);
const raiseLabel = (value: number): string => String(value);
const lastTrickCopy = (): string => "last trick";
const seatPositions = (match: MatchSnapshot): Map<number, string> =>
  new Map(
    match.NumPlayers === 4
      ? [
          [0, "bottom"],
          [1, "right"],
          [2, "top"],
          [3, "left"],
        ]
      : [
          [0, "bottom"],
          [1, "top"],
        ],
  );

test("renderSetupScreen includes offline and online launch areas", () => {
  const html = renderSetupScreen({
    locale: "pt-BR",
    playerName: "Voce",
    relayURL: "",
    transportMode: "",
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    transportOptions: () => `<option value="">auto</option>`,
  });

  assert.match(html, /setup10-offline/);
  assert.match(html, /setup10-online/);
  assert.match(html, /startGame/);
  assert.match(html, /joinOnline/);
});

test("renderLobbyScreen includes invite, seats, and panel tabs", () => {
  const bundle = baseBundle("host_lobby");
  bundle.lobby = {
    invite_key: "ABCD-1234",
    slots: ["Mesa", "Visitante"],
    assigned_seat: 0,
    num_players: 2,
    started: false,
    host_seat: 0,
    connected_seats: { "0": true, "1": true },
    role: "auto",
  };
  bundle.ui.lobby_slots = [
    {
      seat: 0,
      name: "Mesa",
      status: "occupied_online",
      is_empty: false,
      is_local: true,
      is_host: true,
      is_connected: true,
      is_occupied: true,
      is_provisional_cpu: false,
      can_vote_host: false,
      can_request_replacement: false,
    },
    {
      seat: 1,
      name: "Visitante",
      status: "occupied_online",
      is_empty: false,
      is_local: false,
      is_host: false,
      is_connected: true,
      is_occupied: true,
      is_provisional_cpu: false,
      can_vote_host: true,
      can_request_replacement: false,
    },
  ];

  const html = renderLobbyScreen({
    bundle,
    panelTab: "pulse",
    events: [],
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    renderMetric,
    renderEventFeed,
    protocolLabel,
  });

  assert.match(html, /ABCD-1234/);
  assert.match(html, /lobby10-seat-grid/);
  assert.match(html, /data-panel-tab="lobby:pulse"/);
  assert.match(html, /sendHostVote-1/);
});

test("renderGameScreen renders four-player felt table and panel tabs", () => {
  const bundle = baseBundle("offline_match");
  bundle.match = fourPlayerMatch();
  bundle.ui.actions.local_player_id = 0;
  const html = renderGameScreen({
    bundle,
    panelTab: "pulse",
    events: [],
    isOnlineMode: false,
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    renderMetric,
    renderEventFeed,
    renderCard,
    protocolLabel,
    cardLabel,
    playerName,
    teamScore,
    localTeam,
    nextStake,
    raiseLabel,
    lastTrickCopy,
    seatPositions: (match, _bundle) => seatPositions(match),
  });

  assert.match(html, /game10-felt-4/);
  assert.match(html, /game10-seat-top/);
  assert.match(html, /game10-seat-left/);
  assert.match(html, /game10-hand-stage/);
  assert.match(html, /data-panel-tab="game:pulse"/);
});

test("renderGameScreen network tab surfaces failover and seat strip for online play", () => {
  const bundle = baseBundle("host_match");
  bundle.match = fourPlayerMatch();
  bundle.ui.actions.local_player_id = 0;
  bundle.ui.lobby_slots = [
    {
      seat: 0,
      name: "Mesa",
      status: "occupied_online",
      is_empty: false,
      is_local: true,
      is_host: true,
      is_connected: true,
      is_occupied: true,
      is_provisional_cpu: false,
      can_vote_host: false,
      can_request_replacement: false,
    },
    {
      seat: 1,
      name: "Visitante",
      status: "occupied_offline",
      is_empty: false,
      is_local: false,
      is_host: false,
      is_connected: false,
      is_occupied: true,
      is_provisional_cpu: false,
      can_vote_host: true,
      can_request_replacement: true,
    },
  ];
  const html = renderGameScreen({
    bundle,
    panelTab: "network",
    events: [{ kind: "failover_promoted", sequence: 9, timestamp: "2026-04-02T00:00:00Z" } as RuntimeEvent],
    isOnlineMode: true,
    t,
    escapeHtml,
    busyAttr,
    buttonLabel,
    renderMetric,
    renderEventFeed,
    renderCard,
    protocolLabel,
    cardLabel,
    playerName,
    teamScore,
    localTeam,
    nextStake,
    raiseLabel,
    lastTrickCopy,
    seatPositions: (match, _bundle) => seatPositions(match),
  });

  assert.match(html, /signal_failover_promoted/);
  assert.match(html, /game10-seat-strip/);
  assert.match(html, /connection_transport/);
});

function baseBundle(mode: SnapshotBundle["mode"]): SnapshotBundle {
  return {
    versions: { core_api_version: 1, protocol_version: 2, snapshot_schema_version: 2 },
    mode,
    locale: "pt-BR",
    ui: {
      lobby_slots: [],
      actions: {
        local_player_id: 0,
        local_team: 0,
        can_play_card: true,
        can_ask_or_raise: true,
        must_respond: false,
        can_accept: false,
        can_refuse: false,
        can_close_session: true,
      },
    },
    connection: {
      status: mode,
      is_online: mode !== "offline_match",
      is_host: mode.startsWith("host_"),
      network: {
        transport: "relay_quic_v2",
        negotiated_protocol_version: 2,
        supported_protocol_versions: [2],
        seat_protocol_versions: { "0": 2, "1": 2 },
        mixed_protocol_session: false,
      },
      last_event_sequence: 8,
    },
    diagnostics: {
      event_backlog: 0,
      event_log: [],
    },
    lobby: undefined,
    match: undefined,
  };
}

function fourPlayerMatch(): MatchSnapshot {
  return {
    Players: [
      { ID: 0, Name: "Voce", CPU: false, Team: 0, Hand: hand("A", "K", "7") },
      { ID: 1, Name: "CPU-2", CPU: true, Team: 1, Hand: hand("J", "Q", "3") },
      { ID: 2, Name: "CPU-3", CPU: true, Team: 0, Hand: hand("5", "6", "4") },
      { ID: 3, Name: "CPU-4", CPU: true, Team: 1, Hand: hand("2", "10", "9") },
    ],
    NumPlayers: 4,
    CurrentHand: {
      Vira: { Rank: "A", Suit: "Espadas" },
      Manilha: "2",
      Stake: 3,
      TrucoByTeam: 0,
      RaiseRequester: -1,
      Dealer: 0,
      Turn: 0,
      Round: 1,
      RoundStart: 0,
      RoundCards: [
        { PlayerID: 1, Card: { Rank: "7", Suit: "Espadas" }, FaceDown: false },
        { PlayerID: 2, Card: { Rank: "K", Suit: "Espadas" }, FaceDown: false },
      ],
      TrickResults: [-2, -2, -2],
      TrickWins: { "0": 0, "1": 0 },
      WinnerTeam: -1,
      Finished: false,
      PendingRaiseFor: -1,
    },
    LastTrickCards: [],
    TrickPiles: [],
    MatchPoints: { "0": 3, "1": 0 },
    TurnPlayer: 0,
    CurrentTeamTurn: 0,
    Logs: ["Nova mão"],
    WinnerTeam: -1,
    MatchFinished: false,
    CanAskTruco: true,
    PendingRaiseFor: -1,
    PendingRaiseBy: -1,
    PendingRaiseTo: 0,
    CurrentPlayerIdx: 0,
    LastTrickSeq: 0,
    LastTrickTeam: -1,
    LastTrickWinner: -1,
    LastTrickTie: false,
    LastTrickRound: 0,
  };
}

function hand(...ranks: string[]): Card[] {
  return ranks.map((rank, index) => ({
    Rank: rank,
    Suit: ["Espadas", "Copas", "Ouros"][index % 3],
  }));
}
