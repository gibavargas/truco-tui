export type LocaleCode = "pt-BR" | "en-US";

export interface CoreVersions {
  core_api_version: number;
  protocol_version: number;
  snapshot_schema_version: number;
}

export interface Card {
  Rank: string;
  Suit: string;
}

export interface Player {
  ID: number;
  Name: string;
  CPU: boolean;
  ProvisionalCPU?: boolean;
  Team: number;
  Hand: Card[];
}

export interface PlayedCard {
  PlayerID: number;
  Card: Card;
  FaceDown: boolean;
}

export interface TrickPile {
  Winner: number;
  Team: number;
  Round: number;
  Cards: PlayedCard[];
}

export interface HandState {
  Vira: Card;
  Manilha: string;
  Stake: number;
  TrucoByTeam: number;
  RaiseRequester: number;
  Dealer: number;
  Turn: number;
  Round: number;
  RoundStart: number;
  RoundCards: PlayedCard[];
  TrickResults: number[];
  TrickWins: Record<string, number>;
  WinnerTeam: number;
  Finished: boolean;
  PendingRaiseFor: number;
}

export interface MatchSnapshot {
  Players: Player[];
  NumPlayers: number;
  CurrentHand: HandState;
  LastTrickCards: PlayedCard[];
  TrickPiles: TrickPile[];
  MatchPoints: Record<string, number>;
  TurnPlayer: number;
  CurrentTeamTurn: number;
  Logs: string[];
  WinnerTeam: number;
  MatchFinished: boolean;
  CanAskTruco: boolean;
  PendingRaiseFor: number;
  PendingRaiseBy: number;
  PendingRaiseTo: number;
  CurrentPlayerIdx: number;
  LastTrickSeq: number;
  LastTrickTeam: number;
  LastTrickWinner: number;
  LastTrickTie: boolean;
  LastTrickRound: number;
}

export interface LobbySnapshot {
  invite_key?: string;
  slots: string[];
  assigned_seat: number;
  num_players: number;
  started: boolean;
  host_seat: number;
  connected_seats: Record<string, boolean>;
  role?: string;
}

export interface LobbySlotState {
  seat: number;
  name?: string;
  status: string;
  is_empty: boolean;
  is_local: boolean;
  is_host: boolean;
  is_connected: boolean;
  is_occupied: boolean;
  is_provisional_cpu: boolean;
  can_vote_host: boolean;
  can_request_replacement: boolean;
}

export interface ActionSnapshot {
  local_player_id: number;
  local_team: number;
  can_play_card: boolean;
  can_ask_or_raise: boolean;
  must_respond: boolean;
  can_accept: boolean;
  can_refuse: boolean;
  can_close_session: boolean;
}

export interface UIStateSnapshot {
  lobby_slots: LobbySlotState[];
  actions: ActionSnapshot;
}

export interface AppError {
  code: string;
  message: string;
}

export interface NetworkSnapshot {
  transport?: string;
  supported_protocol_versions?: number[];
  negotiated_protocol_version?: number;
  seat_protocol_versions?: Record<string, number>;
  mixed_protocol_session?: boolean;
}

export interface ConnectionSnapshot {
  status: string;
  is_online: boolean;
  is_host: boolean;
  network?: NetworkSnapshot;
  last_error?: AppError;
  last_event_sequence: number;
}

export interface DiagnosticsSnapshot {
  event_backlog: number;
  replay_seed_lo?: number;
  replay_seed_hi?: number;
  event_log?: string[];
}

export interface SnapshotBundle {
  versions: CoreVersions;
  mode: string;
  locale: LocaleCode;
  match?: MatchSnapshot;
  lobby?: LobbySnapshot;
  ui: UIStateSnapshot;
  connection: ConnectionSnapshot;
  diagnostics: DiagnosticsSnapshot;
}

export interface SessionView {
  mode?: string;
  inviteKey?: string;
  numPlayers?: number;
  assignedSeat?: number;
  hostSeat?: number;
  slots?: string[];
  connected?: boolean[];
  started?: boolean;
  role?: string;
  network?: NetworkSnapshot;
}

export interface RuntimeEvent {
  kind: string;
  sequence: number;
  timestamp: string;
  payload?: Record<string, unknown>;
}

export interface ApiResult {
  ok: boolean;
  error?: string;
  error_code?: string;
  sessionId?: string;
  bundle?: SnapshotBundle;
  mode?: string;
  session?: SessionView;
  events?: RuntimeEvent[];
  snapshot?: string;
  sessionClosed?: boolean;
}
