using System.Text.Json;
using System.Text.Json.Serialization;
using TrucoWinUI.Services;

namespace TrucoWinUI.Models;

public static class PlayerHelper
{
    public static Player? GetMe(GameSnapshot? snapshot)
    {
        if (snapshot?.Players == null) return null;
        return snapshot.Players.FirstOrDefault(p => p.ID == snapshot.CurrentPlayerIdx);
    }

    public static Player? GetTopPlayer(GameSnapshot? snapshot)
    {
        if (snapshot?.Players == null) return null;
        return snapshot.NumPlayers == 2
            ? snapshot.Players.FirstOrDefault(p => p.ID == (snapshot.CurrentPlayerIdx + 1) % 2)
            : snapshot.Players.FirstOrDefault(p => p.ID == (snapshot.CurrentPlayerIdx + 2) % 4);
    }

    public static Player? GetRightPlayer(GameSnapshot? snapshot)
    {
        if (snapshot?.NumPlayers != 4 || snapshot?.Players == null) return null;
        return snapshot.Players.FirstOrDefault(p => p.ID == (snapshot.CurrentPlayerIdx + 1) % 4);
    }

    public static Player? GetLeftPlayer(GameSnapshot? snapshot)
    {
        if (snapshot?.NumPlayers != 4 || snapshot?.Players == null) return null;
        return snapshot.Players.FirstOrDefault(p => p.ID == (snapshot.CurrentPlayerIdx + 3) % 4);
    }

    public static int GetMyTeamId(GameSnapshot? snapshot)
    {
        return snapshot?.Players?.FirstOrDefault(p => p.ID == snapshot?.CurrentPlayerIdx)?.Team ?? 0;
    }

    public static bool IsCpuPlayer(GameSnapshot? snapshot, int playerId)
    {
        return snapshot?.Players?.FirstOrDefault(p => p.ID == playerId)?.CPU == true;
    }

    public static string GetPlayerName(GameSnapshot? snapshot, int playerId)
    {
        var player = snapshot?.Players?.FirstOrDefault(p => p.ID == playerId);
        if (player == null) return string.Empty;
        return IsCpuPlayer(snapshot, playerId) 
            ? $"⟳ {player.Name} (CPU)" 
            : $"Vez: {player.Name}";
    }
}

public static class GameStateHelper
{
    public static bool IsPlaying(GameSnapshot? snapshot) => snapshot != null;
    public static bool IsNotPlaying(GameSnapshot? snapshot) => snapshot == null;
    public static bool IsMyTurn(GameSnapshot? snapshot) => snapshot?.TurnPlayer == snapshot?.CurrentPlayerIdx;
    public static bool IsMatchOver(GameSnapshot? snapshot) => snapshot?.MatchFinished == true;

    public static int GetUsPoints(GameSnapshot? snapshot)
    {
        return snapshot?.MatchPoints?.TryGetValue("0", out var u) == true ? u : 0;
    }

    public static int GetThemPoints(GameSnapshot? snapshot)
    {
        return snapshot?.MatchPoints?.TryGetValue("1", out var t) == true ? t : 0;
    }

    public static bool ShowTrucoActions(GameSnapshot? snapshot, int myTeamId)
    {
        return IsPlaying(snapshot) && !IsMatchOver(snapshot) && snapshot?.PendingRaiseFor == myTeamId;
    }

    public static bool ShowAskTruco(GameSnapshot? snapshot)
    {
        return IsPlaying(snapshot) && !IsMatchOver(snapshot) && IsMyTurn(snapshot) && snapshot?.CanAskTruco == true;
    }

    public static string GetMatchResultText(GameSnapshot? snapshot, int myTeamId)
    {
        if (!IsMatchOver(snapshot)) return string.Empty;
        return snapshot?.WinnerTeam == myTeamId ? "VITÓRIA! 🎉" : "DERROTA 😢";
    }

    public static string GetTurnIndicatorText(GameSnapshot? snapshot)
    {
        return IsMyTurn(snapshot) ? "SUA VEZ" : "AGUARDANDO...";
    }

    public static string GetLobbyStatusText(LobbySnapshot? lobby, ConnectionSnapshot? connection)
    {
        if (!string.IsNullOrWhiteSpace(connection?.Status))
        {
            return connection.Status!;
        }

        if (lobby?.Started == true)
        {
            return "partida iniciada";
        }

        var filledSeats = lobby?.Slots?.Count(slot => !string.IsNullOrWhiteSpace(slot)) ?? 0;
        var totalSeats = lobby?.NumPlayers ?? lobby?.Slots?.Count ?? 0;
        return totalSeats > 0 ? $"{filledSeats}/{totalSeats} lugares ocupados" : "aguardando jogadores";
    }

    public static string GetMatchStatusText(GameSnapshot? snapshot, UIStateSnapshot? ui, int myTeamId, IStringProvider strings)
    {
        if (snapshot == null)
        {
            return strings.Get("status.waiting");
        }

        if (snapshot.MatchFinished == true)
        {
            return GetMatchResultText(snapshot, myTeamId);
        }

        if (ui?.Actions?.MustRespond == true)
        {
            return $"Resposta pendente: {GetTrucoLabel(snapshot.PendingRaiseTo)}";
        }

        if (snapshot.PendingRaiseFor == myTeamId)
        {
            return $"Seu time decide {GetTrucoLabel(snapshot.PendingRaiseTo)}";
        }

        if (ui?.Actions?.CanPlayCard == true)
        {
            return strings.Get("turn.yours");
        }

        if (snapshot.TurnPlayer is int turnPlayer)
        {
            return PlayerHelper.GetPlayerName(snapshot, turnPlayer);
        }

        return strings.Get("turn.waiting");
    }

    public static string GetRoundText(GameSnapshot? snapshot)
    {
        return $"Rodada {snapshot?.CurrentHand?.Round ?? 1}/3";
    }

    public static string GetTrucoLabel(int? pendingRaiseTo)
    {
        return pendingRaiseTo switch
        {
            3 => "TRUCO!",
            6 => "SEIS!",
            9 => "NOVE!",
            12 => "DOZE!",
            _ => "TRUCO!"
        };
    }

    public static string GetAskTrucoLabel(int? currentStake)
    {
        return currentStake switch
        {
            1 => "TRUCO!",
            3 => "SEIS!",
            6 => "NOVE!",
            9 => "DOZE!",
            _ => "TRUCO!"
        };
    }

    public static List<(string Label, bool Active)> GetStakeLadder(GameSnapshot? snapshot)
    {
        var stake = snapshot?.CurrentHand?.Stake ?? 1;
        return new List<(string Label, bool Active)>
        {
            ("1", stake >= 1),
            ("T", stake >= 3),
            ("6", stake >= 6),
            ("9", stake >= 9),
            ("12", stake >= 12),
        };
    }

    public static string GetRoleBadge(GameSnapshot? snapshot, int playerId)
    {
        if (snapshot?.CurrentHand == null) return string.Empty;
        var n = snapshot.NumPlayers ?? 2;
        var dealer = snapshot.CurrentHand.Dealer ?? -1;
        if (playerId == dealer) return "🃏";
        if (playerId == (dealer + 1) % n) return "✋";
        if (n == 4 && playerId == (dealer + n - 1) % n) return "🦶";
        return string.Empty;
    }

    public static List<string> GetLogEntries(GameSnapshot? snapshot)
    {
        return snapshot?.Logs?.ToList() ?? new List<string>();
    }
}

public static class JsonOptions
{
    private static JsonSerializerOptions? _default;

    public static JsonSerializerOptions Default
    {
        get
        {
            _default ??= new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true,
                DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
            };
            return _default;
        }
    }
}
