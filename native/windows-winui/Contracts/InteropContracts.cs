using System.Collections.Generic;
using System.Text.Json.Serialization;

namespace TrucoWinUI.Contracts;

public static class RuntimeContract
{
    public const int CoreApiVersion = 1;
    public const int ProtocolVersion = 2;
    public const int SnapshotSchemaVersion = 2;

    public const string SetLocale = "set_locale";
    public const string NewOfflineGame = "new_offline_game";
    public const string NewHand = "new_hand";
    public const string CreateHostSession = "create_host_session";
    public const string JoinSession = "join_session";
    public const string StartHostedMatch = "start_hosted_match";
    public const string GameAction = "game_action";
    public const string Tick = "tick";
    public const string SendChat = "send_chat";
    public const string VoteHost = "vote_host";
    public const string RequestReplacementInvite = "request_replacement_invite";
    public const string CloseSession = "close_session";
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
    [JsonPropertyName("bind_addr")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? BindAddr { get; set; }

    [JsonPropertyName("host_name")]
    public string HostName { get; set; } = "";

    [JsonPropertyName("num_players")]
    public int NumPlayers { get; set; }

    [JsonPropertyName("relay_url")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? RelayUrl { get; set; }

    [JsonPropertyName("transport_mode")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? TransportMode { get; set; }

    [JsonPropertyName("coordinator_url")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? CoordinatorUrl { get; set; }

    [JsonPropertyName("tailnet_control_url")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? TailnetControlUrl { get; set; }
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

    [JsonPropertyName("face_down")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingDefault)]
    public bool FaceDown { get; set; }
}

public sealed class TickIntentPayload
{
    [JsonPropertyName("max_steps")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingDefault)]
    public int MaxSteps { get; set; }
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
