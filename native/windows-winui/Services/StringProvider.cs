namespace TrucoWinUI.Services;

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
            ["turn.cpu"] = "⟳ {0} (CPU)",
            ["round.format"] = "Rodada {0}/3",
            ["truco.label"] = "TRUCO!",
            ["seis.label"] = "SEIS!",
            ["nove.label"] = "NOVE!",
            ["doze.label"] = "DOZE!",
            ["result.victory"] = "VITÓRIA! 🎉",
            ["result.defeat"] = "DERROTA 😢",
            ["player.you"] = "Voce",
            ["player.human"] = "Humano, Time 1",
            ["player.cpu"] = "CPU, Time {0}",
            ["player.cpu.opponent"] = "CPU-Oponente",
            ["player.cpu.right"] = "CPU-Direita",
            ["player.cpu.partner"] = "CPU-Parceiro",
            ["player.cpu.left"] = "CPU-Esquerda",
            ["stake.1"] = "1",
            ["stake.3"] = "T",
            ["stake.6"] = "6",
            ["stake.9"] = "9",
            ["stake.12"] = "12",
            ["game.idle"] = "idle",
            ["team.us"] = "0",
            ["team.them"] = "1"
        },
        ["en-US"] = new Dictionary<string, string>
        {
            ["status.waiting"] = "Runtime waiting for initialization",
            ["status.playing"] = "Running offline game ({0}p)",
            ["turn.yours"] = "YOUR TURN",
            ["turn.waiting"] = "WAITING...",
            ["turn.format"] = "Turn: {0}",
            ["turn.cpu"] = "⟳ {0} (CPU)",
            ["round.format"] = "Round {0}/3",
            ["truco.label"] = "TRUCO!",
            ["seis.label"] = "SEIS!",
            ["nove.label"] = "NOVE!",
            ["doze.label"] = "DOZE!",
            ["result.victory"] = "VICTORY! 🎉",
            ["result.defeat"] = "DEFEAT 😢",
            ["player.you"] = "You",
            ["player.human"] = "Human, Team 1",
            ["player.cpu"] = "CPU, Team {0}",
            ["player.cpu.opponent"] = "CPU-Opponent",
            ["player.cpu.right"] = "CPU-Right",
            ["player.cpu.partner"] = "CPU-Partner",
            ["player.cpu.left"] = "CPU-Left",
            ["stake.1"] = "1",
            ["stake.3"] = "T",
            ["stake.6"] = "6",
            ["stake.9"] = "9",
            ["stake.12"] = "12",
            ["game.idle"] = "idle",
            ["team.us"] = "0",
            ["team.them"] = "1"
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
