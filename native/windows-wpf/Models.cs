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

public class GameSnapshot
{
    public int? NumPlayers { get; set; }
    public Dictionary<string, int>? MatchPoints { get; set; }
    public int? TurnPlayer { get; set; }
    public int? CurrentTeamTurn { get; set; }
    public HandState? CurrentHand { get; set; }
    public List<PlayedCard>? LastTrickCards { get; set; }
    public List<TrickPile>? TrickPiles { get; set; }
    public List<Player>? Players { get; set; }
    public List<string>? Logs { get; set; }
    public int? WinnerTeam { get; set; }
    public bool? MatchFinished { get; set; }
    public bool? CanAskTruco { get; set; }
    public int? PendingRaiseFor { get; set; }
    public int? PendingRaiseBy { get; set; }
    public int? PendingRaiseTo { get; set; }
    public int? CurrentPlayerIdx { get; set; }
    [JsonPropertyName("last_trick_seq")]
    public int? LastTrickSeq { get; set; }

    [JsonPropertyName("last_trick_team")]
    public int? LastTrickTeam { get; set; }

    [JsonPropertyName("last_trick_winner")]
    public int? LastTrickWinner { get; set; }

    [JsonPropertyName("last_trick_tie")]
    public bool? LastTrickTie { get; set; }

    [JsonPropertyName("last_trick_round")]
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
    
    [JsonIgnore]
    public string? WinningCardId
    {
        get
        {
            if (RoundCards == null || RoundCards.Count == 0) return null;
            
            string? bestId = null;
            int bestPower = -1;
            bool isTie = false;

            foreach (var pc in RoundCards)
            {
                if (pc.Card == null) continue;
                int p = pc.Card.Power(Manilha);
                if (p > bestPower)
                {
                    bestPower = p;
                    bestId = $"{pc.PlayerID}-{pc.Card.Rank}-{pc.Card.Suit}";
                    isTie = false;
                }
                else if (p == bestPower)
                {
                    isTie = true;
                }
            }
            return isTie ? null : bestId;
        }
    }
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

public class TrickPile
{
    [JsonPropertyName("Winner")]
    public int? Winner { get; set; }

    [JsonPropertyName("Team")]
    public int? Team { get; set; }

    [JsonPropertyName("Round")]
    public int? Round { get; set; }

    [JsonPropertyName("Cards")]
    public List<PlayedCard>? Cards { get; set; }
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
    
    [JsonIgnore]
    public string Id => $"{Rank}-{Suit}";

    public int Power(string? manilha)
    {
        var normalPower = new Dictionary<string, int>
        {
            { "3", 10 }, { "2", 9 }, { "A", 8 }, { "K", 7 }, { "J", 6 }, 
            { "Q", 5 }, { "7", 4 }, { "6", 3 }, { "5", 2 }, { "4", 1 }
        };
        var manilhaSuitPower = new Dictionary<string, int>
        {
            { "Paus", 4 }, { "Copas", 3 }, { "Espadas", 2 }, { "Ouros", 1 }
        };

        if (Rank == manilha)
        {
            return 100 + (manilhaSuitPower.TryGetValue(Suit, out int sp) ? sp : 0);
        }
        return normalPower.TryGetValue(Rank, out int p) ? p : 0;
    }
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

public class LobbySlotItem
{
    public int Seat { get; set; }
    public string Label { get; set; } = string.Empty;
    public bool IsAssigned { get; set; }
    public bool IsHost { get; set; }
    public bool IsConnected { get; set; }
    public bool IsLocal { get; set; }
    public bool IsProvisionalCpu { get; set; }
    public bool CanVote { get; set; }
    public bool CanReplace { get; set; }
    public string? RuntimeStatus { get; set; }
    public string StatusText
    {
        get
        {
            if (IsProvisionalCpu) return "CPU";
            if (IsHost) return "HOST";
            if (IsLocal) return "VOCE";
            if (!string.IsNullOrWhiteSpace(RuntimeStatus)) return RuntimeStatus!.ToUpperInvariant();
            return IsConnected ? "ONLINE" : (IsAssigned ? "OFFLINE" : "EMPTY");
        }
    }
    public Brush StatusColor => IsProvisionalCpu ? Brushes.Orange : (IsHost ? Brushes.Gold : (IsLocal ? Brushes.DeepSkyBlue : (IsConnected ? Brushes.LightGreen : (IsAssigned ? Brushes.Gray : Brushes.DimGray))));
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

public class AppEvent
{
    [JsonPropertyName("kind")]
    public string? Kind { get; set; }
    [JsonPropertyName("sequence")]
    public int Sequence { get; set; }
    [JsonPropertyName("timestamp")]
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
    [JsonPropertyName("last_error")]
    public AppError? LastError { get; set; }
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
