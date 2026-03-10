using System.Text.Json.Serialization;
using System.Collections.Generic;

namespace TrucoWinUI.Models;

public class GameSnapshot
{
    [JsonPropertyName("mode")]
    public string? Mode { get; set; }
    public int? NumPlayers { get; set; }
    public List<int>? MatchPoints { get; set; }
    public int? TurnPlayer { get; set; }
    public HandState? CurrentHand { get; set; }
    public List<Player>? Players { get; set; }
}

public class HandState
{
    public int? Stake { get; set; }
    public Card? Vira { get; set; }
    public string? Manilha { get; set; }
    public List<PlayedCard>? RoundCards { get; set; }
}

public class Player
{
    public int ID { get; set; }
    public string Name { get; set; } = "";
    public int Team { get; set; }
    public List<Card>? Hand { get; set; }
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
    public string SuitSymbol => Suit switch
    {
        "Espadas" => "♠",
        "Copas" => "♥",
        "Ouros" => "♦",
        "Paus" => "♣",
        _ => ""
    };
}
