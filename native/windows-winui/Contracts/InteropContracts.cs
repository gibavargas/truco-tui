using System.Text.Json.Serialization;

namespace TrucoWinUI.Models;

public sealed class CoreVersions
{
    [JsonPropertyName("core_api_version")]
    public int CoreApiVersion { get; set; }

    [JsonPropertyName("protocol_version")]
    public int ProtocolVersion { get; set; }

    [JsonPropertyName("snapshot_schema_version")]
    public int SnapshotSchemaVersion { get; set; }
}

public sealed class AppIntentEnvelope<TPayload>
{
    [JsonPropertyName("kind")]
    public string Kind { get; set; } = "";

    [JsonPropertyName("payload")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public TPayload? Payload { get; set; }
}

public sealed class SetLocaleIntentPayload
{
    [JsonPropertyName("locale")]
    public string Locale { get; set; } = "";
}

public sealed class NewOfflineGameIntentPayload
{
    [JsonPropertyName("player_names")]
    public List<string> PlayerNames { get; set; } = new();

    [JsonPropertyName("cpu_flags")]
    public List<bool> CpuFlags { get; set; } = new();
}

public sealed class CreateHostSessionIntentPayload
{
    [JsonPropertyName("host_name")]
    public string HostName { get; set; } = "";

    [JsonPropertyName("num_players")]
    public int NumPlayers { get; set; }

    [JsonPropertyName("relay_url")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? RelayUrl { get; set; }
}

public sealed class JoinSessionIntentPayload
{
    [JsonPropertyName("player_name")]
    public string PlayerName { get; set; } = "";

    [JsonPropertyName("key")]
    public string Key { get; set; } = "";

    [JsonPropertyName("desired_role")]
    public string DesiredRole { get; set; } = "";
}

public sealed class GameActionIntentPayload
{
    [JsonPropertyName("action")]
    public string Action { get; set; } = "";

    [JsonPropertyName("card_index")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingDefault)]
    public int CardIndex { get; set; }
}

public sealed class SendChatIntentPayload
{
    [JsonPropertyName("text")]
    public string Text { get; set; } = "";
}

public sealed class HostVoteIntentPayload
{
    [JsonPropertyName("candidate_seat")]
    public int CandidateSeat { get; set; }
}

public sealed class ReplacementInviteIntentPayload
{
    [JsonPropertyName("target_seat")]
    public int TargetSeat { get; set; }
}
