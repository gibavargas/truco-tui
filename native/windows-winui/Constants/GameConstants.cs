namespace TrucoWinUI.Constants;

public static class GameConstants
{
    public const int MinPlayers = 2;
    public const int MaxPlayers = 4;
    public const int DefaultPlayers = 4;

    public const int InitialStake = 1;
    public const int TrucoStake = 3;
    public const int SeisStake = 6;
    public const int NoveStake = 9;
    public const int DozeStake = 12;

    public const int PollIntervalMs = 50;
    public const int DefaultHandCount = 3;

    public const string DefaultPlayerName = "Voce";
    public const string CpuOpponentName = "CPU-Oponente";
    public const string CpuRightName = "CPU-Direita";
    public const string CpuPartnerName = "CPU-Parceiro";
    public const string CpuLeftName = "CPU-Esquerda";

    public static readonly string[] SupportedLocales = { "pt-BR", "en-US" };
    public static readonly string DefaultLocale = "pt-BR";

    public static readonly Dictionary<int, string> StakeLabels = new()
    {
        { 1, "1" },
        { 3, "T" },
        { 6, "6" },
        { 9, "9" },
        { 12, "12" }
    };
}

public static class IntentKinds
{
    public const string SetLocale = "set_locale";
    public const string NewOfflineGame = "new_offline_game";
    public const string GameAction = "game_action";
}

public static class ActionTypes
{
    public const string Play = "play";
    public const string Truco = "truco";
    public const string Accept = "accept";
    public const string Refuse = "refuse";
}

public static class UiConstants
{
    public const string WaitingText = "AGUARDANDO...";
    public const string YourTurnText = "SUA VEZ";
    public const string VictoryText = "VITÓRIA!";
    public const string DefeatText = "DERROTA";
    public const string TrucoLabel = "TRUCO!";
    public const string SeisLabel = "SEIS!";
    public const string NoveLabel = "NOVE!";
    public const string DozeLabel = "DOZE!";
    public const string RoundFormat = "Rodada {0}/3";
    public const string WaitingInitText = "Runtime aguardando inicializacao";
    public const string PlayingFormat = "Rodando partida offline ({0}p)";
    public const string IdleMode = "idle";

    public const string TurnPlayerFormat = "Vez: {0}";
    public const string CpuTurnFormat = "⟳ {0} (CPU)";
    public const string HumanLabel = "Humano, Time 1";
    public const string CpuLabel = "CPU, Time {0}";
    public const string TeamFormat = "Time {0}";
}

public static class SuitSymbols
{
    public const string Espadas = "♠";
    public const string Copas = "♥";
    public const string Ouros = "♦";
    public const string Paus = "♣";
    public const string Empty = "";

    public static readonly Dictionary<string, string> SymbolMap = new()
    {
        { "Espadas", Espadas },
        { "Copas", Copas },
        { "Ouros", Ouros },
        { "Paus", Paus }
    };

    public static string GetSymbol(string suit) => SymbolMap.TryGetValue(suit, out var symbol) ? symbol : Empty;
}

public static class RoleBadges
{
    public const string Dealer = "🃏";
    public const string DealerFirst = "✋";
    public const string DealerLast = "🦶";
    public const string Empty = "";
}
