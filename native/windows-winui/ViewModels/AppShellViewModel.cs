using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using Microsoft.UI.Dispatching;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using TrucoWinUI.Constants;
using TrucoWinUI.Models;
using TrucoWinUI.Services;

namespace TrucoWinUI.ViewModels;

public partial class AppShellViewModel : ObservableObject, IDisposable
{
    private readonly TrucoCoreService _core;
    private readonly DispatcherQueue _dispatcherQueue;
    private readonly IStringProvider _stringProvider;
    private CancellationTokenSource? _pollCts;
    private bool _disposed;

    [ObservableProperty]
    private string status = StringProviderKeys.StatusWaiting;

    [ObservableProperty]
    private string setupPlayerName = GameConstants.DefaultPlayerName;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(SetupPlayerLabels))]
    private int setupNumPlayers = GameConstants.DefaultPlayers;

    [ObservableProperty]
    private int setupLocaleIndex;

    public List<string> SetupPlayerLabels
    {
        get
        {
            var labels = new List<string>();
            var playerName = string.IsNullOrEmpty(SetupPlayerName) 
                ? _stringProvider.Get(StringProviderKeys.PlayerYou) 
                : SetupPlayerName;
            
            for (int i = 0; i < SetupNumPlayers; i++)
            {
                labels.Add(i switch
                {
                    0 => $"{playerName} ({_stringProvider.Get(StringProviderKeys.PlayerHuman)})",
                    1 when SetupNumPlayers == 2 => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuOpponent)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 2)})",
                    1 => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuRight)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 2)})",
                    2 => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuPartner)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 1)})",
                    3 => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuLeft)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 2)})",
                    _ => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuOpponent)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 2)})"
                });
            }
            return labels;
        }
    }

    [ObservableProperty]
    private LobbySnapshot? lobbySnapshot;

    public System.Collections.ObjectModel.ObservableCollection<string> ChatEvents { get; } = new();
    public System.Collections.ObjectModel.ObservableCollection<LobbySlotItem> LobbySlots { get; } = new();

    [ObservableProperty]
    private string inviteKeyInput = "";

    [ObservableProperty]
    private string chatMessage = "";

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(Mode))]
    [NotifyPropertyChangedFor(nameof(IsPlaying))]
    [NotifyPropertyChangedFor(nameof(IsNotPlaying))]
    [NotifyPropertyChangedFor(nameof(IsMyTurn))]
    [NotifyPropertyChangedFor(nameof(UsPoints))]
    [NotifyPropertyChangedFor(nameof(ThemPoints))]
    [NotifyPropertyChangedFor(nameof(ShowTrucoActions))]
    [NotifyPropertyChangedFor(nameof(ShowAskTruco))]
    [NotifyPropertyChangedFor(nameof(IsMatchOver))]
    [NotifyPropertyChangedFor(nameof(MatchResultText))]
    [NotifyPropertyChangedFor(nameof(TrucoLabel))]
    [NotifyPropertyChangedFor(nameof(AskTrucoLabel))]
    [NotifyPropertyChangedFor(nameof(MyTeamID))]
    [NotifyPropertyChangedFor(nameof(TurnIndicatorText))]
    [NotifyPropertyChangedFor(nameof(Me))]
    [NotifyPropertyChangedFor(nameof(TopPlayer))]
    [NotifyPropertyChangedFor(nameof(RightPlayer))]
    [NotifyPropertyChangedFor(nameof(LeftPlayer))]
    [NotifyPropertyChangedFor(nameof(LeftPlayerVisibility))]
    [NotifyPropertyChangedFor(nameof(IsTopPlayerTurn))]
    [NotifyPropertyChangedFor(nameof(IsRightPlayerTurn))]
    [NotifyPropertyChangedFor(nameof(IsLeftPlayerTurn))]
    [NotifyPropertyChangedFor(nameof(RoundText))]
    [NotifyPropertyChangedFor(nameof(StakeLadder))]
    [NotifyPropertyChangedFor(nameof(LogEntries))]
    [NotifyPropertyChangedFor(nameof(IsCpuTurn))]
    [NotifyPropertyChangedFor(nameof(TurnPlayerName))]
    [NotifyPropertyChangedFor(nameof(MyRoleBadge))]
    [NotifyPropertyChangedFor(nameof(TopPlayerRoleBadge))]
    private GameSnapshot? snapshot;

    [ObservableProperty]
    private string mode = UiConstants.IdleMode;

    public bool IsPlaying => GameStateHelper.IsPlaying(Snapshot);
    public bool IsNotPlaying => GameStateHelper.IsNotPlaying(Snapshot);
    public bool IsMyTurn => GameStateHelper.IsMyTurn(Snapshot);

    public Microsoft.UI.Xaml.Visibility VisibilityIfPlaying => (Mode == "offline_match" || Mode == "host_match" || Mode == "client_match" || Mode == "match_over") 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfNotPlaying => (Mode == UiConstants.IdleMode) 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;
        
    public Microsoft.UI.Xaml.Visibility VisibilityIfOnlineLobby => (Mode == "host_lobby" || Mode == "client_lobby") 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfHost => (Mode == "host_lobby") 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfInviteKey => (!string.IsNullOrEmpty(LobbySnapshot?.InviteKey)) 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfMatchOver => IsMatchOver 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfShowTruco => ShowTrucoActions 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public Microsoft.UI.Xaml.Visibility VisibilityIfAskTruco => ShowAskTruco 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public int UsPoints => GameStateHelper.GetUsPoints(Snapshot);
    public int ThemPoints => GameStateHelper.GetThemPoints(Snapshot);

    public int MyTeamID => PlayerHelper.GetMyTeamId(Snapshot);

    public bool ShowTrucoActions => GameStateHelper.ShowTrucoActions(Snapshot, MyTeamID);
    public bool ShowAskTruco => GameStateHelper.ShowAskTruco(Snapshot);
    public bool IsMatchOver => GameStateHelper.IsMatchOver(Snapshot);

    public string MatchResultText => GameStateHelper.GetMatchResultText(Snapshot, MyTeamID);

    public string TurnIndicatorText => GameStateHelper.GetTurnIndicatorText(Snapshot);

    public string RoundText => _stringProvider.Format(StringProviderKeys.RoundFormat, Snapshot?.CurrentHand?.Round ?? 1);

    public string TrucoLabel => GameStateHelper.GetTrucoLabel(Snapshot?.PendingRaiseTo);

    public string AskTrucoLabel => GameStateHelper.GetAskTrucoLabel(Snapshot?.CurrentHand?.Stake);

    public Player? Me => PlayerHelper.GetMe(Snapshot);

    public Player? TopPlayer => PlayerHelper.GetTopPlayer(Snapshot);

    public Player? RightPlayer => PlayerHelper.GetRightPlayer(Snapshot);

    public Player? LeftPlayer => PlayerHelper.GetLeftPlayer(Snapshot);

    public Microsoft.UI.Xaml.Visibility LeftPlayerVisibility => LeftPlayer != null 
        ? Microsoft.UI.Xaml.Visibility.Visible 
        : Microsoft.UI.Xaml.Visibility.Collapsed;

    public bool IsTopPlayerTurn => Snapshot?.TurnPlayer == TopPlayer?.ID;
    public bool IsRightPlayerTurn => Snapshot?.TurnPlayer == RightPlayer?.ID;
    public bool IsLeftPlayerTurn => Snapshot?.TurnPlayer == LeftPlayer?.ID;

    public string MyRoleBadge => GameStateHelper.GetRoleBadge(Snapshot, Me?.ID ?? -1);
    public string TopPlayerRoleBadge => GameStateHelper.GetRoleBadge(Snapshot, TopPlayer?.ID ?? -1);

    public List<(string Label, bool Active)> StakeLadder => GameStateHelper.GetStakeLadder(Snapshot);

    public bool IsCpuTurn => Me != null && PlayerHelper.IsCpuPlayer(Snapshot, Snapshot?.TurnPlayer ?? -1);

    public string TurnPlayerName
    {
        get
        {
            var turnPlayerId = Snapshot?.TurnPlayer ?? -1;
            return PlayerHelper.GetPlayerName(Snapshot, turnPlayerId);
        }
    }

    public List<string> LogEntries => GameStateHelper.GetLogEntries(Snapshot);

    public AppShellViewModel() : this(new TrucoCoreService(), new StringProvider())
    {
    }

    public AppShellViewModel(TrucoCoreService core, IStringProvider stringProvider)
    {
        _core = core ?? throw new ArgumentNullException(nameof(core));
        _stringProvider = stringProvider ?? throw new ArgumentNullException(nameof(stringProvider));
        _dispatcherQueue = DispatcherQueue.GetForCurrentThread();
        
        _pollCts = new CancellationTokenSource();
        _ = PollLoopAsync(_pollCts.Token);
    }

    private async Task PollLoopAsync(CancellationToken ct)
    {
        try
        {
            while (!ct.IsCancellationRequested)
            {
                await Task.Delay(GameConstants.PollIntervalMs, ct);

                var eventJson = _core.PollEventJson();
                if (!string.IsNullOrEmpty(eventJson))
                {
                    try 
                    {
                        var ev = JsonSerializer.Deserialize<AppEvent>(eventJson, JsonOptions.Default);
                        if (ev != null) 
                        {
                            string text = "";
                            if (ev.Kind == "chat") {
                                var author = ev.Payload?.GetProperty("author").GetString() ?? "?";
                                var msg = ev.Payload?.GetProperty("text").GetString() ?? "";
                                text = $"{author}: {msg}";
                            } 
                            else if (ev.Kind == "system") {
                                text = ev.Payload?.GetProperty("text").GetString() ?? "";
                            } 
                            else if (ev.Kind == "replacement_invite") {
                                var link = ev.Payload?.GetProperty("invite_key").GetString() ?? "";
                                text = $"Link de subs: {link}";
                            }
                            if (!string.IsNullOrEmpty(text)) {
                                _dispatcherQueue.TryEnqueue(() => {
                                    ChatEvents.Add(text);
                                });
                            }
                        }
                    } 
                    catch (Exception) { }
                    RefreshSnapshot();
                }
            }
        }
        catch (OperationCanceledException)
        {
        }
        catch (Exception ex)
        {
            Debug.WriteLine($"Poll loop error: {ex.Message}");
        }
    }

    private void RefreshSnapshot()
    {
        var json = _core.SnapshotJson();
        if (string.IsNullOrEmpty(json)) return;

        _dispatcherQueue.TryEnqueue(() =>
        {
            try
            {
                var bundle = JsonSerializer.Deserialize<SnapshotBundle>(json, JsonOptions.Default);
                if (bundle != null)
                {
                    Snapshot = bundle.Match;
                    LobbySnapshot = bundle.Lobby;
                    Mode = bundle.Mode ?? UiConstants.IdleMode;
                    RebuildLobbySlots();
                }
            }
            catch (JsonException ex)
            {
                Debug.WriteLine($"Failed to parse snapshot: {ex.Message}");
            }
            catch (Exception ex)
            {
                Debug.WriteLine($"Unexpected error parsing snapshot: {ex.Message}");
            }
        });
    }

    [RelayCommand]
    private void StartGame()
    {
        try
        {
            var locale = SetupLocaleIndex == 1 ? GameConstants.SupportedLocales[1] : GameConstants.SupportedLocales[0];
            _stringProvider.SetLocale(locale);
            _core.Dispatch($"{{\"kind\":\"{IntentKinds.SetLocale}\",\"payload\":{{\"locale\":\"{locale}\"}}}}");

            var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
            string namesJson;
            string cpusJson;

            if (SetupNumPlayers == GameConstants.MaxPlayers)
            {
                namesJson = $"[\"{name}\",\"{_stringProvider.Get(StringProviderKeys.PlayerCpuRight)}\",\"{_stringProvider.Get(StringProviderKeys.PlayerCpuPartner)}\",\"{_stringProvider.Get(StringProviderKeys.PlayerCpuLeft)}\"]";
                cpusJson = "[false,true,true,true]";
            }
            else
            {
                namesJson = $"[\"{name}\",\"{_stringProvider.Get(StringProviderKeys.PlayerCpuOpponent)}\"]";
                cpusJson = "[false,true]";
            }

            _core.Dispatch($"{{\"kind\":\"{IntentKinds.NewOfflineGame}\",\"payload\":{{\"player_names\":{namesJson},\"cpu_flags\":{cpusJson}}}}}");
            RefreshSnapshot();
            Status = _stringProvider.Format(StringProviderKeys.StatusPlaying, SetupNumPlayers);
        }
        catch (Exception ex)
        {
            Debug.WriteLine($"Failed to start game: {ex.Message}");
            Status = $"Error: {ex.Message}";
        }
    }

    [RelayCommand]
    private void HostOnlineGame()
    {
        var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
        int players = SetupNumPlayers == GameConstants.MaxPlayers ? 4 : 2;
        _core.Dispatch($"{{\"kind\":\"create_host_session\",\"payload\":{{\"host_name\":\"{name}\",\"num_players\":{players}}}}}");
    }

    [RelayCommand]
    private void JoinOnlineGame()
    {
        if (string.IsNullOrEmpty(InviteKeyInput)) return;
        var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
        _core.Dispatch($"{{\"kind\":\"join_session\",\"payload\":{{\"player_name\":\"{name}\",\"key\":\"{InviteKeyInput}\"}}}}");
    }

    [RelayCommand]
    private void SendChat()
    {
        if (string.IsNullOrEmpty(ChatMessage)) return;
        var safeText = ChatMessage.Replace("\"", "\\\"");
        _core.Dispatch($"{{\"kind\":\"send_chat\",\"payload\":{{\"text\":\"{safeText}\"}}}}");
        ChatMessage = "";
    }

    [RelayCommand]
    private void RequestReplacementInvite(object? seatStr)
    {
        if (int.TryParse(seatStr?.ToString(), out int seat))
        {
            _core.Dispatch($"{{\"kind\":\"request_replacement_invite\",\"payload\":{{\"target_seat\":{seat}}}}}");
        }
    }

    [RelayCommand]
    private void VoteHost(object? seatStr)
    {
        if (int.TryParse(seatStr?.ToString(), out int seat))
        {
            _core.Dispatch($"{{\"kind\":\"vote_host\",\"payload\":{{\"candidate_seat\":{seat}}}}}");
        }
    }

    [RelayCommand]
    private void StartHostedMatch()
    {
        _core.Dispatch($"{{\"kind\":\"start_hosted_match\"}}");
    }

    [RelayCommand]
    private void LeaveOnlineGame()
    {
        _core.Dispatch($"{{\"kind\":\"close_session\"}}");
        ChatEvents.Clear();
    }

    [RelayCommand]
    private void BackToSetup()
    {
        Snapshot = null;
        Mode = UiConstants.IdleMode;
    }

    [RelayCommand]
    private void PlayCard(Card? card)
    {
        if (card == null || Me?.Hand == null) return;

        int idx = Me.Hand.FindIndex(c => c.Rank == card.Rank && c.Suit == card.Suit);
        if (idx >= 0)
        {
            _core.Dispatch($"{{\"kind\":\"{IntentKinds.GameAction}\",\"payload\":{{\"action\":\"{ActionTypes.Play}\",\"card_index\":{idx}}}}}");
        }
    }

    [RelayCommand]
    private void RequestTruco()
    {
        _core.Dispatch($"{{\"kind\":\"{IntentKinds.GameAction}\",\"payload\":{{\"action\":\"{ActionTypes.Truco}\"}}}}");
    }

    [RelayCommand]
    private void AcceptTruco()
    {
        _core.Dispatch($"{{\"kind\":\"{IntentKinds.GameAction}\",\"payload\":{{\"action\":\"{ActionTypes.Accept}\"}}}}");
    }

    [RelayCommand]
    private void RefuseTruco()
    {
        _core.Dispatch($"{{\"kind\":\"{IntentKinds.GameAction}\",\"payload\":{{\"action\":\"{ActionTypes.Refuse}\"}}}}");
    }

    public void Dispose()
    {
        Dispose(true);
        GC.SuppressFinalize(this);
    }

    private void Dispose(bool disposing)
    {
        if (_disposed) return;

        if (disposing)
        {
            _pollCts?.Cancel();
            _pollCts?.Dispose();
            _core?.Dispose();
        }

        _disposed = true;
    }

    ~AppShellViewModel()
    {
        Dispose(false);
    }

    private void RebuildLobbySlots()
    {
        LobbySlots.Clear();
        if (LobbySnapshot?.Slots == null) return;

        for (int i = 0; i < LobbySnapshot.Slots.Count; i++)
        {
            var connected = LobbySnapshot.ConnectedSeats?.TryGetValue(i.ToString(), out var isConnected) == true && isConnected;
            LobbySlots.Add(new LobbySlotItem
            {
                Seat = i,
                Label = string.IsNullOrWhiteSpace(LobbySnapshot.Slots[i]) ? "Aguardando..." : LobbySnapshot.Slots[i],
                IsAssigned = LobbySnapshot.AssignedSeat == i,
                IsHost = LobbySnapshot.HostSeat == i,
                IsConnected = connected,
                CanVote = !string.IsNullOrWhiteSpace(LobbySnapshot.Slots[i]) && LobbySnapshot.AssignedSeat != i,
                CanReplace = string.IsNullOrWhiteSpace(LobbySnapshot.Slots[i]) && Mode == "host_lobby",
            });
        }
    }
}

public class LobbySlotItem
{
    public int Seat { get; set; }
    public string Label { get; set; } = "";
    public bool IsAssigned { get; set; }
    public bool IsHost { get; set; }
    public bool IsConnected { get; set; }
    public bool CanVote { get; set; }
    public bool CanReplace { get; set; }
}

public static class StringProviderKeys
{
    public const string StatusWaiting = "status.waiting";
    public const string StatusPlaying = "status.playing";
    public const string TurnYours = "turn.yours";
    public const string TurnWaiting = "turn.waiting";
    public const string TurnFormat = "turn.format";
    public const string TurnCpu = "turn.cpu";
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
}
