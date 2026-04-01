export namespace appcore {
	
	export class ActionSnapshot {
	    local_player_id: number;
	    local_team: number;
	    can_play_card: boolean;
	    can_ask_or_raise: boolean;
	    must_respond: boolean;
	    can_accept: boolean;
	    can_refuse: boolean;
	    can_close_session: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ActionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.local_player_id = source["local_player_id"];
	        this.local_team = source["local_team"];
	        this.can_play_card = source["can_play_card"];
	        this.can_ask_or_raise = source["can_ask_or_raise"];
	        this.must_respond = source["must_respond"];
	        this.can_accept = source["can_accept"];
	        this.can_refuse = source["can_refuse"];
	        this.can_close_session = source["can_close_session"];
	    }
	}
	export class AppError {
	    code: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AppError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class AppEvent {
	    kind: string;
	    sequence: number;
	    timestamp: string;
	    payload?: any;
	
	    static createFrom(source: any = {}) {
	        return new AppEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.sequence = source["sequence"];
	        this.timestamp = source["timestamp"];
	        this.payload = source["payload"];
	    }
	}
	export class NetworkSnapshot {
	    transport?: string;
	    supported_protocol_versions?: number[];
	    negotiated_protocol_version?: number;
	    seat_protocol_versions?: Record<number, number>;
	    mixed_protocol_session?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NetworkSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transport = source["transport"];
	        this.supported_protocol_versions = source["supported_protocol_versions"];
	        this.negotiated_protocol_version = source["negotiated_protocol_version"];
	        this.seat_protocol_versions = source["seat_protocol_versions"];
	        this.mixed_protocol_session = source["mixed_protocol_session"];
	    }
	}
	export class ConnectionSnapshot {
	    status: string;
	    is_online: boolean;
	    is_host: boolean;
	    network?: NetworkSnapshot;
	    last_error?: AppError;
	    last_event_sequence: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.is_online = source["is_online"];
	        this.is_host = source["is_host"];
	        this.network = this.convertValues(source["network"], NetworkSnapshot);
	        this.last_error = this.convertValues(source["last_error"], AppError);
	        this.last_event_sequence = source["last_event_sequence"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CoreVersions {
	    core_api_version: number;
	    protocol_version: number;
	    snapshot_schema_version: number;
	
	    static createFrom(source: any = {}) {
	        return new CoreVersions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.core_api_version = source["core_api_version"];
	        this.protocol_version = source["protocol_version"];
	        this.snapshot_schema_version = source["snapshot_schema_version"];
	    }
	}
	export class DiagnosticsSnapshot {
	    event_backlog: number;
	    replay_seed_lo?: number;
	    replay_seed_hi?: number;
	    event_log?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DiagnosticsSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.event_backlog = source["event_backlog"];
	        this.replay_seed_lo = source["replay_seed_lo"];
	        this.replay_seed_hi = source["replay_seed_hi"];
	        this.event_log = source["event_log"];
	    }
	}
	export class LobbySlotState {
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
	
	    static createFrom(source: any = {}) {
	        return new LobbySlotState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seat = source["seat"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.is_empty = source["is_empty"];
	        this.is_local = source["is_local"];
	        this.is_host = source["is_host"];
	        this.is_connected = source["is_connected"];
	        this.is_occupied = source["is_occupied"];
	        this.is_provisional_cpu = source["is_provisional_cpu"];
	        this.can_vote_host = source["can_vote_host"];
	        this.can_request_replacement = source["can_request_replacement"];
	    }
	}
	export class LobbySnapshot {
	    invite_key?: string;
	    slots?: string[];
	    assigned_seat: number;
	    num_players: number;
	    started: boolean;
	    host_seat: number;
	    connected_seats?: Record<number, boolean>;
	    role?: string;
	    metadata?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new LobbySnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invite_key = source["invite_key"];
	        this.slots = source["slots"];
	        this.assigned_seat = source["assigned_seat"];
	        this.num_players = source["num_players"];
	        this.started = source["started"];
	        this.host_seat = source["host_seat"];
	        this.connected_seats = source["connected_seats"];
	        this.role = source["role"];
	        this.metadata = source["metadata"];
	    }
	}
	
	export class UIStateSnapshot {
	    lobby_slots?: LobbySlotState[];
	    actions: ActionSnapshot;
	
	    static createFrom(source: any = {}) {
	        return new UIStateSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lobby_slots = this.convertValues(source["lobby_slots"], LobbySlotState);
	        this.actions = this.convertValues(source["actions"], ActionSnapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SnapshotBundle {
	    versions: CoreVersions;
	    mode: string;
	    locale: string;
	    match?: truco.Snapshot;
	    lobby?: LobbySnapshot;
	    ui: UIStateSnapshot;
	    connection: ConnectionSnapshot;
	    diagnostics: DiagnosticsSnapshot;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.versions = this.convertValues(source["versions"], CoreVersions);
	        this.mode = source["mode"];
	        this.locale = source["locale"];
	        this.match = this.convertValues(source["match"], truco.Snapshot);
	        this.lobby = this.convertValues(source["lobby"], LobbySnapshot);
	        this.ui = this.convertValues(source["ui"], UIStateSnapshot);
	        this.connection = this.convertValues(source["connection"], ConnectionSnapshot);
	        this.diagnostics = this.convertValues(source["diagnostics"], DiagnosticsSnapshot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace truco {
	
	export class Card {
	    Suit: string;
	    Rank: string;
	
	    static createFrom(source: any = {}) {
	        return new Card(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Suit = source["Suit"];
	        this.Rank = source["Rank"];
	    }
	}
	export class PlayedCard {
	    PlayerID: number;
	    Card: Card;
	    FaceDown: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PlayedCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PlayerID = source["PlayerID"];
	        this.Card = this.convertValues(source["Card"], Card);
	        this.FaceDown = source["FaceDown"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HandState {
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
	    TrickWins: Record<number, number>;
	    WinnerTeam: number;
	    Finished: boolean;
	    PendingRaiseFor: number;
	
	    static createFrom(source: any = {}) {
	        return new HandState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Vira = this.convertValues(source["Vira"], Card);
	        this.Manilha = source["Manilha"];
	        this.Stake = source["Stake"];
	        this.TrucoByTeam = source["TrucoByTeam"];
	        this.RaiseRequester = source["RaiseRequester"];
	        this.Dealer = source["Dealer"];
	        this.Turn = source["Turn"];
	        this.Round = source["Round"];
	        this.RoundStart = source["RoundStart"];
	        this.RoundCards = this.convertValues(source["RoundCards"], PlayedCard);
	        this.TrickResults = source["TrickResults"];
	        this.TrickWins = source["TrickWins"];
	        this.WinnerTeam = source["WinnerTeam"];
	        this.Finished = source["Finished"];
	        this.PendingRaiseFor = source["PendingRaiseFor"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Player {
	    ID: number;
	    Name: string;
	    CPU: boolean;
	    ProvisionalCPU: boolean;
	    Team: number;
	    Hand: Card[];
	    Score: number;
	
	    static createFrom(source: any = {}) {
	        return new Player(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.CPU = source["CPU"];
	        this.ProvisionalCPU = source["ProvisionalCPU"];
	        this.Team = source["Team"];
	        this.Hand = this.convertValues(source["Hand"], Card);
	        this.Score = source["Score"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TrickPile {
	    Winner: number;
	    Team: number;
	    Round: number;
	    Cards: PlayedCard[];
	
	    static createFrom(source: any = {}) {
	        return new TrickPile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Winner = source["Winner"];
	        this.Team = source["Team"];
	        this.Round = source["Round"];
	        this.Cards = this.convertValues(source["Cards"], PlayedCard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snapshot {
	    Players: Player[];
	    NumPlayers: number;
	    CurrentHand: HandState;
	    LastTrickCards: PlayedCard[];
	    TrickPiles: TrickPile[];
	    MatchPoints: Record<number, number>;
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
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Players = this.convertValues(source["Players"], Player);
	        this.NumPlayers = source["NumPlayers"];
	        this.CurrentHand = this.convertValues(source["CurrentHand"], HandState);
	        this.LastTrickCards = this.convertValues(source["LastTrickCards"], PlayedCard);
	        this.TrickPiles = this.convertValues(source["TrickPiles"], TrickPile);
	        this.MatchPoints = source["MatchPoints"];
	        this.TurnPlayer = source["TurnPlayer"];
	        this.CurrentTeamTurn = source["CurrentTeamTurn"];
	        this.Logs = source["Logs"];
	        this.WinnerTeam = source["WinnerTeam"];
	        this.MatchFinished = source["MatchFinished"];
	        this.CanAskTruco = source["CanAskTruco"];
	        this.PendingRaiseFor = source["PendingRaiseFor"];
	        this.PendingRaiseBy = source["PendingRaiseBy"];
	        this.PendingRaiseTo = source["PendingRaiseTo"];
	        this.CurrentPlayerIdx = source["CurrentPlayerIdx"];
	        this.LastTrickSeq = source["LastTrickSeq"];
	        this.LastTrickTeam = source["LastTrickTeam"];
	        this.LastTrickWinner = source["LastTrickWinner"];
	        this.LastTrickTie = source["LastTrickTie"];
	        this.LastTrickRound = source["LastTrickRound"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

