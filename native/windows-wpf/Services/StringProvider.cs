namespace TrucoWPF.Services;

public interface IStringProvider
{
    string Get(string key);
    void SetLocale(string locale);
    string CurrentLocale { get; }
    string Format(string key, params object[] args);
}

public class StringProvider : IStringProvider
{
    private string _currentLocale = "pt-BR";
    private readonly Dictionary<string, Dictionary<string, string>> _resources = new()
    {
        ["pt-BR"] = new Dictionary<string, string>
        {
            ["status.waiting"] = "Runtime aguardando inicializacao",
            ["status.playing"] = "Rodando partida offline ({0}p)",
            ["turn.yours"] = "SUA VEZ",
            ["turn.waiting"] = "AGUARDANDO...",
            ["turn.format"] = "Vez: {0}",
            ["turn.cpu"] = "{0} (CPU)",
            ["round.format"] = "Rodada {0}/3",
            ["truco.label"] = "TRUCO!",
            ["seis.label"] = "SEIS!",
            ["nove.label"] = "NOVE!",
            ["doze.label"] = "DOZE!",
            ["result.victory"] = "VITORIA!",
            ["result.defeat"] = "DERROTA",
            ["player.you"] = "Voce",
            ["player.human"] = "Voce, Time 1",
            ["player.cpu"] = "CPU, Time {0}",
            ["player.cpu.opponent"] = "CPU-Oponente",
            ["player.cpu.right"] = "CPU-Direita",
            ["player.cpu.partner"] = "CPU-Parceiro",
            ["player.cpu.left"] = "CPU-Esquerda",
            ["seat.partner"] = "Parceiro",
            ["seat.opponent"] = "Adversário",
            ["seat.you"] = "Você",
            ["stake.1"] = "1",
            ["stake.3"] = "T",
            ["stake.6"] = "6",
            ["stake.9"] = "9",
            ["stake.12"] = "12",
            ["game.idle"] = "idle",
            ["game.playing"] = "playing",
            ["team.us"] = "Nos",
            ["team.them"] = "Eles"
        },
        ["en-US"] = new Dictionary<string, string>
        {
            ["status.waiting"] = "Runtime waiting for initialization",
            ["status.playing"] = "Running offline game ({0}p)",
            ["turn.yours"] = "YOUR TURN",
            ["turn.waiting"] = "WAITING...",
            ["turn.format"] = "Turn: {0}",
            ["turn.cpu"] = "{0} (CPU)",
            ["round.format"] = "Round {0}/3",
            ["truco.label"] = "TRUCO!",
            ["seis.label"] = "SEIS!",
            ["nove.label"] = "NOVE!",
            ["doze.label"] = "DOZE!",
            ["result.victory"] = "VICTORY!",
            ["result.defeat"] = "DEFEAT",
            ["player.you"] = "You",
            ["player.human"] = "You, Team 1",
            ["player.cpu"] = "CPU, Team {0}",
            ["player.cpu.opponent"] = "CPU-Opponent",
            ["player.cpu.right"] = "CPU-Right",
            ["player.cpu.partner"] = "CPU-Partner",
            ["player.cpu.left"] = "CPU-Left",
            ["seat.partner"] = "Partner",
            ["seat.opponent"] = "Opponent",
            ["seat.you"] = "You",
            ["stake.1"] = "1",
            ["stake.3"] = "T",
            ["stake.6"] = "6",
            ["stake.9"] = "9",
            ["stake.12"] = "12",
            ["game.idle"] = "idle",
            ["game.playing"] = "playing",
            ["team.us"] = "Us",
            ["team.them"] = "Them"
        }
    };

    public string CurrentLocale => _currentLocale;

    public string Get(string key)
    {
        if (_resources.TryGetValue(_currentLocale, out var localeResources))
        {
            if (localeResources.TryGetValue(key, out var value))
            {
                return value;
            }
        }
        if (_resources["pt-BR"].TryGetValue(key, out var fallback))
        {
            return fallback;
        }
        return key;
    }

    public void SetLocale(string locale)
    {
        if (_resources.ContainsKey(locale))
        {
            _currentLocale = locale;
        }
    }

    public string Format(string key, params object[] args)
    {
        return string.Format(Get(key), args);
    }
}

public static class StringProviderKeys
{
    public const string StatusWaiting = "status.waiting";
    public const string StatusPlaying = "status.playing";
    public const string TurnYours = "turn.yours";
    public const string TurnWaiting = "turn.waiting";
    public const string RoundFormat = "round.format";
    public const string TrucoLabel = "truco.label";
    public const string SeisLabel = "seis.label";
    public const string NoveLabel = "nove.label";
    public const string DozeLabel = "doze.label";
    public const string ResultVictory = "result.victory";
    public const string ResultDefeat = "result.defeat";
    public const string PlayerYou = "player.you";
    public const string PlayerHuman = "player.human";
    public const string PlayerCpu = "player.cpu";
    public const string PlayerCpuOpponent = "player.cpu.opponent";
    public const string PlayerCpuRight = "player.cpu.right";
    public const string PlayerCpuPartner = "player.cpu.partner";
    public const string PlayerCpuLeft = "player.cpu.left";
    public const string SeatPartner = "seat.partner";
    public const string SeatOpponent = "seat.opponent";
    public const string SeatYou = "seat.you";
    public const string TeamUs = "team.us";
    public const string TeamThem = "team.them";
}
