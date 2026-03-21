using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace TrucoWinUI.Models;

public sealed class AppError
{
    [JsonPropertyName("code")]
    public string Code { get; set; } = string.Empty;

    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;

    public override string ToString() => $"{Code}: {Message}";
}

public sealed class CoreVersions
{
    [JsonPropertyName("core_api_version")]
    public int CoreApiVersion { get; set; }

    [JsonPropertyName("protocol_version")]
    public int ProtocolVersion { get; set; }

    [JsonPropertyName("snapshot_schema_version")]
    public int SnapshotSchemaVersion { get; set; }
}

public sealed class LobbySnapshot
{
    [JsonPropertyName("invite_key")]
    public string? InviteKey { get; set; }

    [JsonPropertyName("slots")]
    public List<string> Slots { get; set; } = [];

    [JsonPropertyName("assigned_seat")]
    public int AssignedSeat { get; set; }

    [JsonPropertyName("num_players")]
    public int NumPlayers { get; set; }

    [JsonPropertyName("started")]
    public bool Started { get; set; }

    [JsonPropertyName("host_seat")]
    public int HostSeat { get; set; }

    [JsonPropertyName("connected_seats")]
    public Dictionary<int, bool> ConnectedSeats { get; set; } = [];

    [JsonPropertyName("role")]
    public string? Role { get; set; }

    [JsonPropertyName("metadata")]
    public Dictionary<string, JsonElement> Metadata { get; set; } = [];
}

public sealed class ConnectionSnapshot
{
    [JsonPropertyName("status")]
    public string Status { get; set; } = "idle";

    [JsonPropertyName("is_online")]
    public bool IsOnline { get; set; }

    [JsonPropertyName("is_host")]
    public bool IsHost { get; set; }

    [JsonPropertyName("network")]
    public NetworkSnapshot? Network { get; set; }

    [JsonPropertyName("last_error")]
    public AppError? LastError { get; set; }

    [JsonPropertyName("last_event_sequence")]
    public long LastEventSequence { get; set; }
}

public sealed class DiagnosticsSnapshot
{
    [JsonPropertyName("event_backlog")]
    public int EventBacklog { get; set; }

    [JsonPropertyName("replay_seed_lo")]
    public ulong ReplaySeedLo { get; set; }

    [JsonPropertyName("replay_seed_hi")]
    public ulong ReplaySeedHi { get; set; }

    [JsonPropertyName("event_log")]
    public List<string> EventLog { get; set; } = [];
}

public sealed class NetworkSnapshot
{
    [JsonPropertyName("transport")]
    public string Transport { get; set; } = string.Empty;

    [JsonPropertyName("supported_protocol_versions")]
    public List<int> SupportedProtocolVersions { get; set; } = [];

    [JsonPropertyName("negotiated_protocol_version")]
    public int NegotiatedProtocolVersion { get; set; }

    [JsonPropertyName("seat_protocol_versions")]
    public Dictionary<int, int> SeatProtocolVersions { get; set; } = [];

    [JsonPropertyName("mixed_protocol_session")]
    public bool MixedProtocolSession { get; set; }
}

public sealed class SnapshotBundle
{
    [JsonPropertyName("versions")]
    public CoreVersions Versions { get; set; } = new();

    [JsonPropertyName("mode")]
    public string Mode { get; set; } = "idle";

    [JsonPropertyName("locale")]
    public string Locale { get; set; } = "pt-BR";

    [JsonPropertyName("match")]
    public MatchSnapshot? Match { get; set; }

    [JsonPropertyName("lobby")]
    public LobbySnapshot? Lobby { get; set; }

    [JsonPropertyName("connection")]
    public ConnectionSnapshot Connection { get; set; } = new();

    [JsonPropertyName("diagnostics")]
    public DiagnosticsSnapshot Diagnostics { get; set; } = new();
}

public sealed class AppEvent
{
    [JsonPropertyName("kind")]
    public string Kind { get; set; } = string.Empty;

    [JsonPropertyName("sequence")]
    public long Sequence { get; set; }

    [JsonPropertyName("timestamp")]
    public string Timestamp { get; set; } = string.Empty;

    [JsonPropertyName("payload")]
    public JsonElement Payload { get; set; }
}

public sealed class MatchSnapshot
{
    [JsonPropertyName("Players")]
    public List<PlayerState> Players { get; set; } = [];

    [JsonPropertyName("NumPlayers")]
    public int NumPlayers { get; set; }

    [JsonPropertyName("CurrentHand")]
    public HandState CurrentHand { get; set; } = new();

    [JsonPropertyName("MatchPoints")]
    public Dictionary<int, int> MatchPoints { get; set; } = [];

    [JsonPropertyName("TurnPlayer")]
    public int TurnPlayer { get; set; }

    [JsonPropertyName("CurrentTeamTurn")]
    public int CurrentTeamTurn { get; set; }

    [JsonPropertyName("Logs")]
    public List<string> Logs { get; set; } = [];

    [JsonPropertyName("WinnerTeam")]
    public int WinnerTeam { get; set; } = -1;

    [JsonPropertyName("MatchFinished")]
    public bool MatchFinished { get; set; }

    [JsonPropertyName("CanAskTruco")]
    public bool CanAskTruco { get; set; }

    [JsonPropertyName("PendingRaiseFor")]
    public int PendingRaiseFor { get; set; } = -1;

    [JsonPropertyName("PendingRaiseBy")]
    public int PendingRaiseBy { get; set; } = -1;

    [JsonPropertyName("PendingRaiseTo")]
    public int PendingRaiseTo { get; set; }

    [JsonPropertyName("CurrentPlayerIdx")]
    public int CurrentPlayerIdx { get; set; } = -1;

    [JsonPropertyName("LastTrickSeq")]
    public int LastTrickSeq { get; set; }

    [JsonPropertyName("LastTrickTeam")]
    public int LastTrickTeam { get; set; } = -1;

    [JsonPropertyName("LastTrickWinner")]
    public int LastTrickWinner { get; set; } = -1;

    [JsonPropertyName("LastTrickTie")]
    public bool LastTrickTie { get; set; }

    [JsonPropertyName("LastTrickRound")]
    public int LastTrickRound { get; set; }
}

public sealed class HandState
{
    [JsonPropertyName("Vira")]
    public CardState Vira { get; set; } = new();

    [JsonPropertyName("Manilha")]
    public string Manilha { get; set; } = string.Empty;

    [JsonPropertyName("Stake")]
    public int Stake { get; set; }

    [JsonPropertyName("TrucoByTeam")]
    public int TrucoByTeam { get; set; } = -1;

    [JsonPropertyName("RaiseRequester")]
    public int RaiseRequester { get; set; } = -1;

    [JsonPropertyName("Dealer")]
    public int Dealer { get; set; }

    [JsonPropertyName("Turn")]
    public int Turn { get; set; }

    [JsonPropertyName("Round")]
    public int Round { get; set; }

    [JsonPropertyName("RoundStart")]
    public int RoundStart { get; set; }

    [JsonPropertyName("RoundCards")]
    public List<PlayedCardState> RoundCards { get; set; } = [];

    [JsonPropertyName("TrickResults")]
    public List<int> TrickResults { get; set; } = [];

    [JsonPropertyName("TrickWins")]
    public Dictionary<int, int> TrickWins { get; set; } = [];

    [JsonPropertyName("WinnerTeam")]
    public int WinnerTeam { get; set; } = -1;

    [JsonPropertyName("Finished")]
    public bool Finished { get; set; }

    [JsonPropertyName("PendingRaiseFor")]
    public int PendingRaiseFor { get; set; } = -1;
}

public sealed class PlayerState
{
    [JsonPropertyName("ID")]
    public int Id { get; set; }

    [JsonPropertyName("Name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("CPU")]
    public bool Cpu { get; set; }

    [JsonPropertyName("ProvisionalCPU")]
    public bool ProvisionalCpu { get; set; }

    [JsonPropertyName("Team")]
    public int Team { get; set; }

    [JsonPropertyName("Hand")]
    public List<CardState> Hand { get; set; } = [];

    [JsonPropertyName("Score")]
    public int Score { get; set; }
}

public sealed class PlayedCardState
{
    [JsonPropertyName("PlayerID")]
    public int PlayerId { get; set; }

    [JsonPropertyName("Card")]
    public CardState Card { get; set; } = new();
}

public sealed class CardState
{
    [JsonPropertyName("Rank")]
    public string Rank { get; set; } = string.Empty;

    [JsonPropertyName("Suit")]
    public string Suit { get; set; } = string.Empty;

    [JsonIgnore]
    public string SuitSymbol => Suit switch
    {
        "Espadas" => "\u2660",
        "Copas" => "\u2665",
        "Ouros" => "\u2666",
        "Paus" => "\u2663",
        _ => string.Empty,
    };

    [JsonIgnore]
    public string ShortLabel => $"{Rank}{SuitSymbol}";

    [JsonIgnore]
    public bool IsRed => Suit is "Copas" or "Ouros";
}

public sealed class LobbySeatViewModel
{
    public int SeatIndex { get; set; }
    public string Name { get; set; } = string.Empty;
    public bool IsAssigned { get; set; }
    public bool IsConnected { get; set; }
    public bool IsHost { get; set; }
    public bool IsEmpty { get; set; }
    public int ProtocolVersion { get; set; }
    public string StatusText { get; set; } = string.Empty;
    public string DisplayLabel => $"Slot {SeatIndex + 1}: {Name}";
    public string ConnectionBadge => IsConnected ? "Conectado" : IsEmpty ? "Livre" : "Offline";
    public string RoleBadge => IsHost ? "Host" : IsAssigned ? "Local" : string.Empty;
    public string ProtocolBadge => ProtocolVersion > 0 ? $"v{ProtocolVersion}" : string.Empty;
}

public sealed class TableSeatViewModel
{
    public int SeatIndex { get; set; } = -1;
    public int PlayerId { get; set; } = -1;
    public string Name { get; set; } = string.Empty;
    public string RoleLabel { get; set; } = string.Empty;
    public int TeamIndex { get; set; } = -1;
    public string TeamLabel { get; set; } = string.Empty;
    public bool IsVisible { get; set; }
    public bool IsLocal { get; set; }
    public bool IsCurrentTurn { get; set; }
    public bool IsCpu { get; set; }
    public bool IsProvisionalCpu { get; set; }
    public int HandCount { get; set; }
    public List<HandCardViewModel> HandCards { get; set; } = [];
    public CardState? PlayedCard { get; set; }
    public string Summary => IsVisible ? $"{Name}  {TeamLabel}" : string.Empty;
    public string CpuTag => IsProvisionalCpu ? "CPU temporaria" : IsCpu ? "CPU" : string.Empty;
    public string PlayedCardLabel => PlayedCard?.ShortLabel ?? "--";
    public string HiddenHandText => HandCount <= 0 ? string.Empty : string.Join(" ", Enumerable.Repeat("[ ]", Math.Min(HandCount, 3)));
}

public sealed class HandCardViewModel
{
    public CardState? Card { get; set; }
    public bool IsFaceUp { get; set; }
    public double Rotation { get; set; }
    public double Scale { get; set; } = 1.0;
    public double TranslateX { get; set; }
    public double TranslateY { get; set; }
    public Microsoft.UI.Xaml.Visibility CardVisibility => Card is null ? Microsoft.UI.Xaml.Visibility.Collapsed : Microsoft.UI.Xaml.Visibility.Visible;
}

public sealed class ActivityEntry
{
    public string Channel { get; set; } = "system";
    public string Timestamp { get; set; } = string.Empty;
    public string Text { get; set; } = string.Empty;
    public string Accent { get; set; } = "#E8E8E8";

    public string ChannelLabel => Channel switch
    {
        "chat" => "CHAT",
        "system" => "SISTEMA",
        "error" => "ERRO",
        "match" => "JOGO",
        _ => Channel.ToUpperInvariant(),
    };

    public override string ToString()
        => string.IsNullOrWhiteSpace(Timestamp)
            ? $"{Channel}: {Text}"
            : $"{Timestamp} {Channel}: {Text}";
}

public static class CollectionExtensions
{
    public static void ReplaceWith<T>(this ObservableCollection<T> target, IEnumerable<T> items)
    {
        target.Clear();
        foreach (T item in items)
        {
            target.Add(item);
        }
    }
}
