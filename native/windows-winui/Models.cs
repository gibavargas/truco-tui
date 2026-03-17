using System.Text.Json.Serialization;
using System.Collections.Generic;

namespace TrucoWinUI.Models;

// Top-level bundle from FFI
public class SnapshotBundle
{
    [JsonPropertyName("versions")]
    public CoreVersions? Versions { get; set; }

    [JsonPropertyName("mode")]
    public string? Mode { get; set; }
    
    [JsonPropertyName("locale")]
    public string? Locale { get; set; }
    
    [JsonPropertyName("match")]
    public GameSnapshot? Match { get; set; }
    
    [JsonPropertyName("lobby")]
    public LobbySnapshot? Lobby { get; set; }

    [JsonPropertyName("ui")]
    public UIStateSnapshot? UI { get; set; }
    
    [JsonPropertyName("connection")]
    public ConnectionSnapshot? Connection { get; set; }
    
    [JsonPropertyName("diagnostics")]
    public DiagnosticsSnapshot? Diagnostics { get; set; }
}

public class UIStateSnapshot
{
    [JsonPropertyName("lobby_slots")]
    public List<LobbySlotState>? LobbySlots { get; set; }

    [JsonPropertyName("actions")]
    public ActionSnapshot? Actions { get; set; }
}

public class LobbySlotState
{
    [JsonPropertyName("seat")]
    public int Seat { get; set; }

    [JsonPropertyName("name")]
    public string? Name { get; set; }

    [JsonPropertyName("status")]
    public string? Status { get; set; }

    [JsonPropertyName("is_empty")]
    public bool IsEmpty { get; set; }

    [JsonPropertyName("is_local")]
    public bool IsLocal { get; set; }

    [JsonPropertyName("is_host")]
    public bool IsHost { get; set; }

    [JsonPropertyName("is_connected")]
    public bool IsConnected { get; set; }

    [JsonPropertyName("is_occupied")]
    public bool IsOccupied { get; set; }

    [JsonPropertyName("is_provisional_cpu")]
    public bool IsProvisionalCpu { get; set; }

    [JsonPropertyName("can_vote_host")]
    public bool CanVoteHost { get; set; }

    [JsonPropertyName("can_request_replacement")]
    public bool CanRequestReplacement { get; set; }
}

public class ActionSnapshot
{
    [JsonPropertyName("local_player_id")]
    public int LocalPlayerId { get; set; }

    [JsonPropertyName("local_team")]
    public int LocalTeam { get; set; }

    [JsonPropertyName("can_play_card")]
    public bool CanPlayCard { get; set; }

    [JsonPropertyName("can_ask_or_raise")]
    public bool CanAskOrRaise { get; set; }

    [JsonPropertyName("must_respond")]
    public bool MustRespond { get; set; }

    [JsonPropertyName("can_accept")]
    public bool CanAccept { get; set; }

    [JsonPropertyName("can_refuse")]
    public bool CanRefuse { get; set; }

    [JsonPropertyName("can_close_session")]
    public bool CanCloseSession { get; set; }
}

public class AppEvent
{
    [JsonPropertyName("kind")]
    public string Kind { get; set; } = "";
    
    [JsonPropertyName("sequence")]
    public long Sequence { get; set; }
    
    [JsonPropertyName("timestamp")]
    public string Timestamp { get; set; } = "";
    
    [JsonPropertyName("payload")]
    public JsonElement? Payload { get; set; }
}

public class GameSnapshot
{
    public int? NumPlayers { get; set; }
    public Dictionary<string, int>? MatchPoints { get; set; }
    public int? TurnPlayer { get; set; }
    public int? CurrentTeamTurn { get; set; }
    public HandState? CurrentHand { get; set; }
    public List<Player>? Players { get; set; }
    public List<string>? Logs { get; set; }
    public int? WinnerTeam { get; set; }
    public bool? MatchFinished { get; set; }
    public bool? CanAskTruco { get; set; }
    public int? PendingRaiseFor { get; set; }
    public int? PendingRaiseBy { get; set; }
    public int? PendingRaiseTo { get; set; }
    public int? CurrentPlayerIdx { get; set; }
    public int? LastTrickSeq { get; set; }
    public int? LastTrickTeam { get; set; }
    public int? LastTrickWinner { get; set; }
    public bool? LastTrickTie { get; set; }
    public int? LastTrickRound { get; set; }
}

public class HandState
{
    public int? Stake { get; set; }
    public Card? Vira { get; set; }
    public string? Manilha { get; set; }
    public List<PlayedCard>? RoundCards { get; set; }
    public List<int>? TrickResults { get; set; }
    public Dictionary<string, int>? TrickWins { get; set; }
    public int? Round { get; set; }
    public int? RoundStart { get; set; }
    public int? Dealer { get; set; }
    public int? Turn { get; set; }
    public int? TrucoByTeam { get; set; }
    public int? RaiseRequester { get; set; }
    public int? WinnerTeam { get; set; }
    public bool? Finished { get; set; }
    public int? PendingRaiseFor { get; set; }
}

public class Player
{
    public int ID { get; set; }
    public string Name { get; set; } = "";
    public int Team { get; set; }
    public List<Card>? Hand { get; set; }
    public bool? CPU { get; set; }
    public bool? ProvisionalCPU { get; set; }
}

public class PlayedCard
{
    public int PlayerID { get; set; }
    public Card? Card { get; set; }
}

public class Card
{
    public string Rank { get; set; } = "";
    public string Suit { get; set; } = "";

    [JsonIgnore]
    public bool IsRed => Suit == "Copas" || Suit == "Ouros";

    [JsonIgnore]
    public Microsoft.UI.Xaml.Media.SolidColorBrush ColorBrush => IsRed 
        ? new Microsoft.UI.Xaml.Media.SolidColorBrush(Microsoft.UI.Colors.Red) 
        : new Microsoft.UI.Xaml.Media.SolidColorBrush(Microsoft.UI.Colors.Black);

    [JsonIgnore]
    public string SuitSymbol => Suit switch
    {
        "Espadas" => "♠",
        "Copas" => "♥",
        "Ouros" => "♦",
        "Paus" => "♣",
        _ => ""
    };
}

public class LobbySnapshot
{
    [JsonPropertyName("invite_key")]
    public string? InviteKey { get; set; }
    [JsonPropertyName("slots")]
    public List<string>? Slots { get; set; }
    [JsonPropertyName("assigned_seat")]
    public int? AssignedSeat { get; set; }
    [JsonPropertyName("num_players")]
    public int? NumPlayers { get; set; }
    [JsonPropertyName("started")]
    public bool? Started { get; set; }
    [JsonPropertyName("host_seat")]
    public int? HostSeat { get; set; }
    [JsonPropertyName("connected_seats")]
    public Dictionary<string, bool>? ConnectedSeats { get; set; }
    [JsonPropertyName("role")]
    public string? Role { get; set; }
}

public class ConnectionSnapshot
{
    [JsonPropertyName("status")]
    public string? Status { get; set; }
    [JsonPropertyName("is_online")]
    public bool? IsOnline { get; set; }
    [JsonPropertyName("is_host")]
    public bool? IsHost { get; set; }
    [JsonPropertyName("last_error")]
    public AppError? LastError { get; set; }

    [JsonPropertyName("last_event_sequence")]
    public long LastEventSequence { get; set; }
}

public class DiagnosticsSnapshot
{
    [JsonPropertyName("event_backlog")]
    public int? EventBacklog { get; set; }
    [JsonPropertyName("event_log")]
    public List<string>? EventLog { get; set; }
}

public class AppError
{
    [JsonPropertyName("code")]
    public string? Code { get; set; }

    [JsonPropertyName("message")]
    public string? Message { get; set; }
}
