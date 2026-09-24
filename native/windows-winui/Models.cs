using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;
using CommunityToolkit.Mvvm.ComponentModel;

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

    [JsonPropertyName("requested_transport")]
    public string RequestedTransport { get; set; } = string.Empty;

    [JsonPropertyName("direct_path_known")]
    public bool DirectPathKnown { get; set; }

    [JsonPropertyName("direct_path")]
    public bool DirectPath { get; set; }

    [JsonPropertyName("relay_fallback")]
    public bool RelayFallback { get; set; }

    [JsonPropertyName("coordinator_status")]
    public string CoordinatorStatus { get; set; } = string.Empty;

    [JsonPropertyName("coordinator_url")]
    public string CoordinatorUrl { get; set; } = string.Empty;

    [JsonPropertyName("tailnet_node")]
    public string TailnetNode { get; set; } = string.Empty;

    [JsonPropertyName("tailnet_authority")]
    public string TailnetAuthority { get; set; } = string.Empty;

    [JsonPropertyName("tailnet_service_port")]
    public int TailnetServicePort { get; set; }

    [JsonPropertyName("fallback_reason")]
    public string FallbackReason { get; set; } = string.Empty;

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

    [JsonPropertyName("ui")]
    public UIStateSnapshot Ui { get; set; } = new();

    [JsonPropertyName("connection")]
    public ConnectionSnapshot Connection { get; set; } = new();

    [JsonPropertyName("diagnostics")]
    public DiagnosticsSnapshot Diagnostics { get; set; } = new();
}

public sealed class UIStateSnapshot
{
    [JsonPropertyName("lobby_slots")]
    public List<LobbySlotState> LobbySlots { get; set; } = [];

    [JsonPropertyName("actions")]
    public ActionSnapshot Actions { get; set; } = new();
}

public sealed class LobbySlotState
{
    [JsonPropertyName("seat")]
    public int Seat { get; set; }

    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

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

public sealed class ActionSnapshot
{
    [JsonPropertyName("local_player_id")]
    public int LocalPlayerId { get; set; } = -1;

    [JsonPropertyName("local_team")]
    public int LocalTeam { get; set; } = -1;

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

    [JsonPropertyName("LastTrickCards")]
    public List<PlayedCardState> LastTrickCards { get; set; } = [];

    [JsonPropertyName("TrickPiles")]
    public List<TrickPileState> TrickPiles { get; set; } = [];

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

    [JsonPropertyName("FaceDown")]
    public bool FaceDown { get; set; }
}

public sealed class TrickPileState
{
    [JsonPropertyName("Winner")]
    public int Winner { get; set; } = -1;

    [JsonPropertyName("Team")]
    public int Team { get; set; } = -1;

    [JsonPropertyName("Round")]
    public int Round { get; set; }

    [JsonPropertyName("Cards")]
    public List<PlayedCardState> Cards { get; set; } = [];
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

    [JsonIgnore]
    public string AccessibilityLabel
    {
        get
        {
            if (string.IsNullOrWhiteSpace(Rank) && string.IsNullOrWhiteSpace(Suit))
            {
                return "Carta";
            }

            if (string.IsNullOrWhiteSpace(Suit))
            {
                return Rank;
            }

            return $"{Rank} de {Suit}";
        }
    }
}

public sealed class LobbySeatViewModel
{
    public int SeatIndex { get; set; }
    public string Name { get; set; } = string.Empty;
    public bool IsAssigned { get; set; }
    public bool IsConnected { get; set; }
    public bool IsHost { get; set; }
    public bool IsEmpty { get; set; }
    public bool IsProvisionalCpu { get; set; }
    public bool CanVoteHost { get; set; }
    public bool CanRequestReplacement { get; set; }
    public int ProtocolVersion { get; set; }
    public string StatusText { get; set; } = string.Empty;
    public string DisplayLabel => $"Slot {SeatIndex + 1}: {Name}";
    public string ConnectionBadge => IsProvisionalCpu ? "CPU provisória" : IsConnected ? "Conectado" : IsEmpty ? "Livre" : "Offline";
    public string RoleBadge => IsHost ? "Host" : IsAssigned ? "Local" : string.Empty;
    public string ProtocolBadge => ProtocolVersion > 0 ? $"v{ProtocolVersion}" : string.Empty;
    public bool HasRoleBadge => !string.IsNullOrWhiteSpace(RoleBadge);
    public bool HasProtocolBadge => !string.IsNullOrWhiteSpace(ProtocolBadge);
    public string ActionHint => CanRequestReplacement
        ? "Gera um convite para substituir uma CPU provisoria ou trazer de volta um jogador desconectado."
        : CanVoteHost
            ? "Vota neste assento para assumir o host se o host atual cair."
            : IsProvisionalCpu
                ? "Este assento esta ocupado por uma CPU provisoria ate um substituto entrar."
                : "Nenhuma acao especial disponivel neste assento agora.";
    public string AccessibilityLabel
    {
        get
        {
            List<string> parts =
            [
                DisplayLabel,
                StatusText,
                ConnectionBadge,
            ];
            if (HasRoleBadge)
            {
                parts.Add(RoleBadge);
            }

            if (HasProtocolBadge)
            {
                parts.Add($"Compatibilidade {ProtocolBadge}");
            }

            return string.Join(". ", parts);
        }
    }
}

public sealed class TableSeatViewModel : ObservableObject
{
    private int seatIndex = -1;
    private int playerId = -1;
    private string name = string.Empty;
    private string roleLabel = string.Empty;
    private int teamIndex = -1;
    private string teamLabel = string.Empty;
    private bool isVisible;
    private bool isLocal;
    private bool isCurrentTurn;
    private bool isCpu;
    private bool isProvisionalCpu;
    private int handCount;
    private ObservableCollection<HandCardViewModel> handCards = [];
    private CardState? playedCard;
    private HandCardViewModel? playedCardViewModel;

    public int SeatIndex { get => seatIndex; set => SetProperty(ref seatIndex, value); }
    public int PlayerId { get => playerId; set => SetProperty(ref playerId, value); }
    public string Name
    {
        get => name;
        set
        {
            if (SetProperty(ref name, value))
            {
                OnPropertyChanged(nameof(Summary));
                OnPropertyChanged(nameof(AccessibilityLabel));
                OnPropertyChanged(nameof(PlayedCardAccessibilityLabel));
                OnPropertyChanged(nameof(HiddenHandAccessibilityLabel));
            }
        }
    }
    public string RoleLabel
    {
        get => roleLabel;
        set
        {
            if (SetProperty(ref roleLabel, value))
            {
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public int TeamIndex { get => teamIndex; set => SetProperty(ref teamIndex, value); }
    public string TeamLabel
    {
        get => teamLabel;
        set
        {
            if (SetProperty(ref teamLabel, value))
            {
                OnPropertyChanged(nameof(Summary));
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public bool IsVisible
    {
        get => isVisible;
        set
        {
            if (SetProperty(ref isVisible, value))
            {
                OnPropertyChanged(nameof(Summary));
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public bool IsLocal { get => isLocal; set => SetProperty(ref isLocal, value); }
    public bool IsCurrentTurn
    {
        get => isCurrentTurn;
        set
        {
            if (SetProperty(ref isCurrentTurn, value))
            {
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public bool IsCpu
    {
        get => isCpu;
        set
        {
            if (SetProperty(ref isCpu, value))
            {
                OnPropertyChanged(nameof(CpuTag));
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public bool IsProvisionalCpu
    {
        get => isProvisionalCpu;
        set
        {
            if (SetProperty(ref isProvisionalCpu, value))
            {
                OnPropertyChanged(nameof(CpuTag));
                OnPropertyChanged(nameof(AccessibilityLabel));
            }
        }
    }
    public int HandCount
    {
        get => handCount;
        set
        {
            if (SetProperty(ref handCount, value))
            {
                OnPropertyChanged(nameof(HiddenHandText));
                OnPropertyChanged(nameof(HiddenHandAccessibilityLabel));
            }
        }
    }
    public ObservableCollection<HandCardViewModel> HandCards { get => handCards; set => SetProperty(ref handCards, value); }
    public CardState? PlayedCard
    {
        get => playedCard;
        set
        {
            if (SetProperty(ref playedCard, value))
            {
                OnPropertyChanged(nameof(PlayedCardLabel));
                OnPropertyChanged(nameof(PlayedCardAccessibilityLabel));
            }
        }
    }
    public HandCardViewModel? PlayedCardViewModel { get => playedCardViewModel; set => SetProperty(ref playedCardViewModel, value); }

    public string Summary => IsVisible ? $"{Name}  {TeamLabel}" : string.Empty;
    public string CpuTag => IsProvisionalCpu ? "CPU temporaria" : IsCpu ? "CPU" : string.Empty;
    public string PlayedCardLabel => PlayedCard?.ShortLabel ?? "--";
    public string HiddenHandText => HandCount <= 0 ? string.Empty : string.Join(" ", Enumerable.Repeat("[ ]", Math.Min(HandCount, 3)));
    public string AccessibilityLabel
    {
        get
        {
            if (!IsVisible)
            {
                return string.Empty;
            }

            List<string> parts =
            [
                string.IsNullOrWhiteSpace(RoleLabel) ? Name : $"{RoleLabel}: {Name}",
                TeamLabel,
            ];

            if (!string.IsNullOrWhiteSpace(CpuTag))
            {
                parts.Add(CpuTag);
            }

            parts.Add(IsCurrentTurn ? "Com a vez" : "Aguardando");
            return string.Join(". ", parts);
        }
    }
    public string PlayedCardAccessibilityLabel => PlayedCard is null
        ? $"{Name} ainda nao jogou carta."
        : $"{Name} jogou {PlayedCard.AccessibilityLabel}.";
    public string HiddenHandAccessibilityLabel => HandCount <= 0
        ? $"{Name} nao tem cartas na mao."
        : $"{Name} tem {HandCount} cartas na mao.";
}

public sealed class HandCardViewModel : ObservableObject
{
    private CardState? card;
    private bool isFaceUp;
    private double rotation;
    private double scale = 1.0;
    private double translateX;
    private double translateY;

    public CardState? Card
    {
        get => card;
        set
        {
            if (SetProperty(ref card, value))
            {
                OnPropertyChanged(nameof(CardVisibility));
                OnPropertyChanged(nameof(AccessibilityLabel));
                OnPropertyChanged(nameof(PlayAutomationLabel));
                OnPropertyChanged(nameof(FaceDownAutomationLabel));
            }
        }
    }

    public bool IsFaceUp
    {
        get => isFaceUp;
        set
        {
            if (SetProperty(ref isFaceUp, value))
            {
                OnPropertyChanged(nameof(AccessibilityLabel));
                OnPropertyChanged(nameof(PlayAutomationLabel));
                OnPropertyChanged(nameof(FaceDownAutomationLabel));
            }
        }
    }

    public double Rotation { get => rotation; set => SetProperty(ref rotation, value); }
    public double Scale { get => scale; set => SetProperty(ref scale, value); }
    public double TranslateX { get => translateX; set => SetProperty(ref translateX, value); }
    public double TranslateY { get => translateY; set => SetProperty(ref translateY, value); }
    public Microsoft.UI.Xaml.Visibility CardVisibility => Card is null ? Microsoft.UI.Xaml.Visibility.Collapsed : Microsoft.UI.Xaml.Visibility.Visible;
    public string AccessibilityLabel => Card is null
        ? "Carta indisponivel"
        : IsFaceUp
            ? Card.AccessibilityLabel
            : "Carta virada";
    public string PlayAutomationLabel => Card is null
        ? "Carta indisponivel"
        : IsFaceUp
            ? $"Jogar {Card.AccessibilityLabel}"
            : "Jogar carta";
    public string FaceDownAutomationLabel => Card is null
        ? "Carta indisponivel"
        : $"Jogar {Card.AccessibilityLabel} virada";
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
