using System.Text.Json.Serialization;
using System.Collections.Generic;
using System.Windows.Media;

namespace TrucoWPF.Models;

public class SnapshotBundle
{
    [JsonPropertyName("mode")]
    public string? Mode { get; set; }
    
    [JsonPropertyName("locale")]
    public string? Locale { get; set; }
    
    [JsonPropertyName("match")]
    public GameSnapshot? Match { get; set; }
    
    [JsonPropertyName("lobby")]
    public LobbySnapshot? Lobby { get; set; }
    
    [JsonPropertyName("connection")]
    public ConnectionSnapshot? Connection { get; set; }
    
    [JsonPropertyName("diagnostics")]
    public DiagnosticsSnapshot? Diagnostics { get; set; }
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
    public Brush ColorBrush => IsRed ? Brushes.Red : Brushes.Black;

    [JsonIgnore]
    public string SuitSymbol => Suit switch
    {
        "Espadas" => "♠",
        "Copas" => "♥",
        "Ouros" => "♦",
        "Paus" => "♣",
        _ => ""
    };

    [JsonIgnore]
    public string DisplayText => $"{Rank}{SuitSymbol}";
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
    public List<int>? ConnectedSeats { get; set; }
}

public class LobbySlotItem
{
    public int Seat { get; set; }
    public string Label { get; set; } = string.Empty;
    public bool IsAssigned { get; set; }
    public bool IsHost { get; set; }
    public bool IsConnected { get; set; }
    public bool CanVote { get; set; }
    public bool CanReplace { get; set; }
    public string StatusText => IsHost ? "HOST" : (IsConnected ? "ONLINE" : (IsAssigned ? "OFFLINE" : "EMPTY"));
    public Brush StatusColor => IsHost ? Brushes.Gold : (IsConnected ? Brushes.LightGreen : (IsAssigned ? Brushes.Gray : Brushes.DimGray));
}

public class AppEvent
{
    [JsonPropertyName("kind")]
    public string? Kind { get; set; }
    [JsonPropertyName("seq")]
    public int Seq { get; set; }
    [JsonPropertyName("ts")]
    public string? Timestamp { get; set; }
    [JsonPropertyName("payload")]
    public System.Text.Json.JsonElement? Payload { get; set; }
}

public class ConnectionSnapshot
{
    [JsonPropertyName("status")]
    public string? Status { get; set; }
    [JsonPropertyName("is_online")]
    public bool? IsOnline { get; set; }
    [JsonPropertyName("is_host")]
    public bool? IsHost { get; set; }
}

public class DiagnosticsSnapshot
{
    [JsonPropertyName("event_backlog")]
    public int? EventBacklog { get; set; }
    [JsonPropertyName("event_log")]
    public List<string>? EventLog { get; set; }
}
