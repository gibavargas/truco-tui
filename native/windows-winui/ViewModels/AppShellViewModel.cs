using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using Microsoft.UI.Dispatching;
using Microsoft.UI.Xaml;
using System.Collections.ObjectModel;
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
using Windows.ApplicationModel.DataTransfer;

namespace TrucoWinUI.ViewModels;

public partial class AppShellViewModel : ObservableObject, IDisposable
{
    private readonly TrucoCoreService _core;
    private readonly DispatcherQueue _dispatcherQueue;
    private readonly IStringProvider _stringProvider;
    private CancellationTokenSource? _pollCts;
    private string? _lastSnapshotJson;
    private bool _disposed;

    [ObservableProperty]
    private string status = StringProviderKeys.StatusWaiting;

    [ObservableProperty]
    private string setupPlayerName = GameConstants.DefaultPlayerName;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(SetupPlayerLabels))]
    private int setupNumPlayers = GameConstants.DefaultPlayers;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(SetupPlayerLabels))]
    [NotifyPropertyChangedFor(nameof(SetupSelectedPlayerCount))]
    private int setupNumPlayersIndex = 1;

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
            
            for (int i = 0; i < SetupSelectedPlayerCount; i++)
            {
                labels.Add(i switch
                {
                    0 => $"{playerName} ({_stringProvider.Get(StringProviderKeys.PlayerHuman)})",
                    1 when SetupSelectedPlayerCount == 2 => $"{_stringProvider.Get(StringProviderKeys.PlayerCpuOpponent)} ({string.Format(_stringProvider.Get(StringProviderKeys.PlayerCpu), 2)})",
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
    [NotifyPropertyChangedFor(nameof(VisibilityIfInviteKey))]
    [NotifyPropertyChangedFor(nameof(InviteKeyText))]
    [NotifyPropertyChangedFor(nameof(LobbyStatusText))]
    [NotifyPropertyChangedFor(nameof(ConnectionRoleText))]
    private LobbySnapshot? lobbySnapshot;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(IsMyTurn))]
    [NotifyPropertyChangedFor(nameof(ShowTrucoActions))]
    [NotifyPropertyChangedFor(nameof(ShowAskTruco))]
    [NotifyPropertyChangedFor(nameof(CanPlayCards))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfOnlineMatch))]
    [NotifyPropertyChangedFor(nameof(CanCloseSession))]
    [NotifyPropertyChangedFor(nameof(MatchStatusText))]
    private UIStateSnapshot? uiState;

    public System.Collections.ObjectModel.ObservableCollection<string> ChatEvents { get; } = new();
    public System.Collections.ObjectModel.ObservableCollection<LobbySlotItem> LobbySlots { get; } = new();

    [ObservableProperty]
    private string inviteKeyInput = "";

    [ObservableProperty]
    private string chatMessage = "";

    [ObservableProperty]
    private string setupRelayUrl = "";

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(SetupDesiredRole))]
    private int setupDesiredRoleIndex;

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
    [NotifyPropertyChangedFor(nameof(MatchStatusText))]
    private GameSnapshot? snapshot;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(VisibilityIfPlaying))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfNotPlaying))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfOnlineLobby))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfOnlineMatch))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfHost))]
    private string mode = UiConstants.IdleMode;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(ConnectionStatusText))]
    [NotifyPropertyChangedFor(nameof(ConnectionModeText))]
    [NotifyPropertyChangedFor(nameof(ConnectionRoleText))]
    [NotifyPropertyChangedFor(nameof(ConnectionErrorText))]
    [NotifyPropertyChangedFor(nameof(EventBacklogText))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfConnectionRole))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfConnectionError))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfCombinedError))]
    [NotifyPropertyChangedFor(nameof(LobbyStatusText))]
    [NotifyPropertyChangedFor(nameof(CombinedErrorText))]
    private ConnectionSnapshot? connectionState;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(EventBacklogText))]
    private DiagnosticsSnapshot? diagnosticsState;

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(VisibilityIfLastActionError))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfCombinedError))]
    [NotifyPropertyChangedFor(nameof(CombinedErrorText))]
    private string lastActionError = "";

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(VisibilityIfLastActionError))]
    [NotifyPropertyChangedFor(nameof(VisibilityIfCombinedError))]
    [NotifyPropertyChangedFor(nameof(CombinedErrorText))]
    private string lastActionErrorCode = "";

    public bool IsPlaying => GameStateHelper.IsPlaying(Snapshot);
    public bool IsNotPlaying => GameStateHelper.IsNotPlaying(Snapshot);
    public bool IsMyTurn => UiState?.Actions?.CanPlayCard == true || UiState?.Actions?.MustRespond == true;

    public ObservableCollection<string> MatchLogEntries { get; } = new();
    public ObservableCollection<string> EventLogEntries { get; } = new();
    public ObservableCollection<string> LobbyEventEntries { get; } = new();

    public Visibility VisibilityIfPlaying => (Mode == "offline_match" || Mode == "host_match" || Mode == "client_match" || Mode == "match_over")
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfNotPlaying => (Mode == UiConstants.IdleMode)
        ? Visibility.Visible
        : Visibility.Collapsed;
        
    public Visibility VisibilityIfOnlineLobby => (Mode == "host_lobby" || Mode == "client_lobby")
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfOnlineMatch => (Mode == "host_match" || Mode == "client_match")
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfHost => (Mode == "host_lobby")
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfInviteKey => !string.IsNullOrEmpty(LobbySnapshot?.InviteKey)
        ? Visibility.Visible
        : Visibility.Collapsed;
    public Visibility VisibilityIfConnectionRole => !string.IsNullOrWhiteSpace(ConnectionRoleText)
        ? Visibility.Visible
        : Visibility.Collapsed;
    public Visibility VisibilityIfConnectionError => !string.IsNullOrWhiteSpace(ConnectionErrorText)
        ? Visibility.Visible
        : Visibility.Collapsed;
    public Visibility VisibilityIfLastActionError => !string.IsNullOrWhiteSpace(LastActionError)
        ? Visibility.Visible
        : Visibility.Collapsed;
    public Visibility VisibilityIfCombinedError => !string.IsNullOrWhiteSpace(CombinedErrorText)
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfMatchOver => IsMatchOver
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfShowTruco => ShowTrucoActions
        ? Visibility.Visible
        : Visibility.Collapsed;

    public Visibility VisibilityIfAskTruco => ShowAskTruco
        ? Visibility.Visible
        : Visibility.Collapsed;

    public int UsPoints => GameStateHelper.GetUsPoints(Snapshot);
    public int ThemPoints => GameStateHelper.GetThemPoints(Snapshot);

    public int MyTeamID => PlayerHelper.GetMyTeamId(Snapshot);

    public bool ShowTrucoActions => UiState?.Actions?.MustRespond == true;
    public bool ShowAskTruco => UiState?.Actions?.CanAskOrRaise == true && UiState?.Actions?.MustRespond != true;
    public bool IsMatchOver => GameStateHelper.IsMatchOver(Snapshot);
    public bool CanPlayCards => UiState?.Actions?.CanPlayCard == true;
    public bool CanCloseSession => UiState?.Actions?.CanCloseSession == true;
    public int SetupSelectedPlayerCount => SetupNumPlayersIndex == 1 ? 4 : 2;
    public string SetupDesiredRole => SetupDesiredRoleIndex switch
    {
        1 => "partner",
        2 => "opponent",
        _ => "auto"
    };
    public string ConnectionStatusText => ConnectionState?.Status ?? Mode;
    public string ConnectionModeText => ConnectionState?.IsOnline == true ? "online" : "offline";
    public string ConnectionRoleText => LobbySnapshot?.Role ?? SetupDesiredRole;
    public string ConnectionErrorText => ConnectionState?.LastError?.Message ?? "";
    public string EventBacklogText => $"{DiagnosticsState?.EventBacklog ?? 0}";
    public string InviteKeyText => string.IsNullOrWhiteSpace(LobbySnapshot?.InviteKey) ? "-" : LobbySnapshot!.InviteKey!;
    public string MatchStatusText => GameStateHelper.GetMatchStatusText(Snapshot, UiState, MyTeamID, _stringProvider);
    public string LobbyStatusText => GameStateHelper.GetLobbyStatusText(LobbySnapshot, ConnectionState);
    public string CombinedErrorText => !string.IsNullOrWhiteSpace(LastActionError)
        ? (string.IsNullOrWhiteSpace(LastActionErrorCode) ? LastActionError : $"{LastActionErrorCode}: {LastActionError}")
        : ConnectionErrorText;

    public string MatchResultText => GameStateHelper.GetMatchResultText(Snapshot, MyTeamID);

    public string TurnIndicatorText => GameStateHelper.GetTurnIndicatorText(Snapshot);

    public string RoundText => _stringProvider.Format(StringProviderKeys.RoundFormat, Snapshot?.CurrentHand?.Round ?? 1);

    public string TrucoLabel => GameStateHelper.GetTrucoLabel(Snapshot?.PendingRaiseTo);

    public string AskTrucoLabel => GameStateHelper.GetAskTrucoLabel(Snapshot?.CurrentHand?.Stake);

    public Player? Me => PlayerHelper.GetMe(Snapshot);

    public Player? TopPlayer => PlayerHelper.GetTopPlayer(Snapshot);

    public Player? RightPlayer => PlayerHelper.GetRightPlayer(Snapshot);

    public Player? LeftPlayer => PlayerHelper.GetLeftPlayer(Snapshot);

    public Visibility LeftPlayerVisibility => LeftPlayer != null
        ? Visibility.Visible
        : Visibility.Collapsed;

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
        RefreshSnapshot();
    }

    private async Task PollLoopAsync(CancellationToken ct)
    {
        try
        {
            while (!ct.IsCancellationRequested)
            {
                await Task.Delay(GameConstants.PollIntervalMs, ct);

                var snapshotJson = _core.SnapshotJson();
                if (!string.IsNullOrWhiteSpace(snapshotJson) && !string.Equals(snapshotJson, _lastSnapshotJson, StringComparison.Ordinal))
                {
                    _lastSnapshotJson = snapshotJson;
                    ApplySnapshotJson(snapshotJson);
                }

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
                                    TrimCollection(ChatEvents, 80);
                                });
                            }
                            _dispatcherQueue.TryEnqueue(() => AppendEvent(ev));
                        }
                    } 
                    catch (Exception) { }
                    var latestSnapshot = _core.SnapshotJson();
                    if (!string.IsNullOrWhiteSpace(latestSnapshot))
                    {
                        _lastSnapshotJson = latestSnapshot;
                        ApplySnapshotJson(latestSnapshot);
                    }
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
        _lastSnapshotJson = json;
        ApplySnapshotJson(json);
    }

    private void ApplySnapshotJson(string json)
    {
        _dispatcherQueue.TryEnqueue(() =>
        {
            try
            {
                var bundle = JsonSerializer.Deserialize<SnapshotBundle>(json, JsonOptions.Default);
                if (bundle != null)
                {
                    Snapshot = bundle.Match;
                    LobbySnapshot = bundle.Lobby;
                    UiState = bundle.UI;
                    ConnectionState = bundle.Connection;
                    DiagnosticsState = bundle.Diagnostics;
                    Mode = bundle.Mode ?? UiConstants.IdleMode;
                    if (!string.IsNullOrWhiteSpace(bundle.Locale))
                    {
                        _stringProvider.SetLocale(bundle.Locale!);
                    }
                    RebuildLobbySlots();
                    RebuildLogs();
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
            LastActionError = "";
            SetLocaleFromSetup();

            var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
            var payload = new NewOfflineGameIntentPayload();

            if (SetupSelectedPlayerCount == GameConstants.MaxPlayers)
            {
                payload.PlayerNames.AddRange(new[]
                {
                    name,
                    _stringProvider.Get(StringProviderKeys.PlayerCpuRight),
                    _stringProvider.Get(StringProviderKeys.PlayerCpuPartner),
                    _stringProvider.Get(StringProviderKeys.PlayerCpuLeft),
                });
                payload.CpuFlags.AddRange(new[] { false, true, true, true });
            }
            else
            {
                payload.PlayerNames.AddRange(new[]
                {
                    name,
                    _stringProvider.Get(StringProviderKeys.PlayerCpuOpponent),
                });
                payload.CpuFlags.AddRange(new[] { false, true });
            }

            DispatchIntent(IntentKinds.NewOfflineGame, payload);
            Status = _stringProvider.Format(StringProviderKeys.StatusPlaying, SetupSelectedPlayerCount);
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
        LastActionError = "";
        var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
        SetLocaleFromSetup();
        DispatchIntent("create_host_session", new CreateHostSessionIntentPayload
        {
            HostName = name,
            NumPlayers = SetupSelectedPlayerCount,
            RelayUrl = string.IsNullOrWhiteSpace(SetupRelayUrl) ? null : SetupRelayUrl.Trim(),
        });
    }

    [RelayCommand]
    private void JoinOnlineGame()
    {
        if (string.IsNullOrWhiteSpace(InviteKeyInput)) return;
        LastActionError = "";
        var name = string.IsNullOrEmpty(SetupPlayerName) ? GameConstants.DefaultPlayerName : SetupPlayerName;
        SetLocaleFromSetup();
        DispatchIntent("join_session", new JoinSessionIntentPayload
        {
            PlayerName = name,
            Key = InviteKeyInput.Trim(),
            DesiredRole = SetupDesiredRole,
        });
    }

    [RelayCommand]
    private void SendChat()
    {
        if (string.IsNullOrWhiteSpace(ChatMessage)) return;
        DispatchIntent("send_chat", new SendChatIntentPayload
        {
            Text = ChatMessage.Trim(),
        });
        ChatMessage = "";
    }

    [RelayCommand]
    private void RequestReplacementInvite(object? seatStr)
    {
        if (int.TryParse(seatStr?.ToString(), out int seat))
        {
            DispatchIntent("request_replacement_invite", new ReplacementInviteIntentPayload
            {
                TargetSeat = seat,
            });
        }
    }

    [RelayCommand]
    private void VoteHost(object? seatStr)
    {
        if (int.TryParse(seatStr?.ToString(), out int seat))
        {
            DispatchIntent("vote_host", new HostVoteIntentPayload
            {
                CandidateSeat = seat,
            });
        }
    }

    [RelayCommand]
    private void StartHostedMatch()
    {
        DispatchIntent<object?>("start_hosted_match", null);
    }

    [RelayCommand]
    private void LeaveOnlineGame()
    {
        if (!CanCloseSession)
        {
            return;
        }

        DispatchIntent<object?>("close_session", null);
        ChatEvents.Clear();
        LobbyEventEntries.Clear();
        EventLogEntries.Clear();
        MatchLogEntries.Clear();
    }

    [RelayCommand]
    private void LeaveMatch()
    {
        if (!CanCloseSession)
        {
            return;
        }

        DispatchIntent<object?>("close_session", null);
        ChatEvents.Clear();
        LobbyEventEntries.Clear();
        EventLogEntries.Clear();
        MatchLogEntries.Clear();
    }

    [RelayCommand]
    private void RefreshState()
    {
        RefreshSnapshot();
    }

    [RelayCommand]
    private void CopyInviteKey()
    {
        if (string.IsNullOrWhiteSpace(LobbySnapshot?.InviteKey))
        {
            return;
        }

        var data = new DataPackage();
        data.SetText(LobbySnapshot.InviteKey);
        Clipboard.SetContent(data);
    }

    [RelayCommand]
    private void BackToSetup()
    {
        if (!CanCloseSession)
        {
            return;
        }

        DispatchIntent<object?>("close_session", null);
        Snapshot = null;
        LobbySnapshot = null;
        UiState = null;
        ConnectionState = null;
        DiagnosticsState = null;
        Mode = UiConstants.IdleMode;
        ChatEvents.Clear();
        LobbySlots.Clear();
        MatchLogEntries.Clear();
        EventLogEntries.Clear();
        LobbyEventEntries.Clear();
    }

    [RelayCommand]
    private void PlayCard(Card? card)
    {
        if (card == null || Me?.Hand == null) return;

        int idx = Me.Hand.FindIndex(c => c.Rank == card.Rank && c.Suit == card.Suit);
        if (idx >= 0)
        {
            DispatchIntent(IntentKinds.GameAction, new GameActionIntentPayload
            {
                Action = ActionTypes.Play,
                CardIndex = idx,
            });
        }
    }

    [RelayCommand]
    private void RequestTruco()
    {
        DispatchIntent(IntentKinds.GameAction, new GameActionIntentPayload
        {
            Action = ActionTypes.Truco,
        });
    }

    [RelayCommand]
    private void AcceptTruco()
    {
        DispatchIntent(IntentKinds.GameAction, new GameActionIntentPayload
        {
            Action = ActionTypes.Accept,
        });
    }

    [RelayCommand]
    private void RefuseTruco()
    {
        DispatchIntent(IntentKinds.GameAction, new GameActionIntentPayload
        {
            Action = ActionTypes.Refuse,
        });
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
        if (UiState?.LobbySlots != null && UiState.LobbySlots.Count > 0)
        {
            foreach (var slot in UiState.LobbySlots)
            {
                LobbySlots.Add(new LobbySlotItem
                {
                    Seat = slot.Seat,
                    Label = string.IsNullOrWhiteSpace(slot.Name) ? "Aguardando..." : slot.Name!,
                    IsAssigned = slot.IsOccupied,
                    IsHost = slot.IsHost,
                    IsConnected = slot.IsConnected,
                    IsLocal = slot.IsLocal,
                    IsProvisionalCpu = slot.IsProvisionalCpu,
                    RuntimeStatus = slot.Status,
                    CanVote = slot.CanVoteHost,
                    CanReplace = slot.CanRequestReplacement,
                });
            }
        }
    }

    private void RebuildLogs()
    {
        ReplaceCollection(MatchLogEntries, Snapshot?.Logs);
        ReplaceCollection(EventLogEntries, DiagnosticsState?.EventLog);
    }

    private void ReplaceCollection(ObservableCollection<string> target, IEnumerable<string>? source)
    {
        target.Clear();
        if (source == null)
        {
            return;
        }

        foreach (var item in source.TakeLast(80))
        {
            target.Add(item);
        }
    }

    private void SetLocaleFromSetup()
    {
        var locale = SetupLocaleIndex == 1 ? GameConstants.SupportedLocales[1] : GameConstants.SupportedLocales[0];
        _stringProvider.SetLocale(locale);
        DispatchIntent(IntentKinds.SetLocale, new SetLocaleIntentPayload
        {
            Locale = locale,
        }, refresh: false);
    }

    private void DispatchIntent<TPayload>(string kind, TPayload? payload, bool refresh = true)
    {
        try
        {
            var response = _core.Dispatch(JsonSerializer.Serialize(new AppIntentEnvelope<TPayload>
            {
                Kind = kind,
                Payload = payload,
            }, JsonOptions.Default));
            CaptureActionError(response);
            if (refresh)
            {
                RefreshSnapshot();
            }
        }
        catch (Exception ex)
        {
            LastActionError = ex.Message;
            LastActionErrorCode = "";
            Debug.WriteLine($"Dispatch failed: {ex.Message}");
        }
    }

    private void CaptureActionError(string? response)
    {
        if (string.IsNullOrWhiteSpace(response))
        {
            LastActionError = "";
            LastActionErrorCode = "";
            return;
        }

        try
        {
            var error = JsonSerializer.Deserialize<AppError>(response, JsonOptions.Default);
            if (!string.IsNullOrWhiteSpace(error?.Message))
            {
                LastActionError = error.Message ?? "";
                LastActionErrorCode = error.Code ?? "";
                return;
            }
        }
        catch (JsonException)
        {
        }

        LastActionError = "";
        LastActionErrorCode = "";
    }

    private void AppendEvent(AppEvent ev)
    {
        var line = FormatEventLine(ev);
        if (string.IsNullOrWhiteSpace(line))
        {
            return;
        }

        EventLogEntries.Add(line);
        TrimCollection(EventLogEntries, 80);

        if (Mode == "host_lobby" || Mode == "client_lobby")
        {
            LobbyEventEntries.Add(line);
            TrimCollection(LobbyEventEntries, 80);
        }
    }

    private string FormatEventLine(AppEvent ev)
    {
        var stamp = FormatTimestamp(ev.Timestamp);
        var payload = ev.Payload;
        return ev.Kind switch
        {
            "chat" => $"{stamp}{ReadPayloadString(payload, "author", "?")}: {ReadPayloadString(payload, "text")}",
            "system" => $"{stamp}{ReadPayloadString(payload, "text", "system")}",
            "replacement_invite" => $"{stamp}invite: {ReadPayloadString(payload, "invite_key")}",
            "error" => $"{stamp}error: {ReadPayloadString(payload, "message")}",
            "lobby_updated" => $"{stamp}lobby updated",
            "match_updated" => $"{stamp}match updated",
            _ => ""
        };
    }

    private static string ReadPayloadString(JsonElement? payload, string property, string fallback = "")
    {
        if (payload is JsonElement element &&
            element.ValueKind == JsonValueKind.Object &&
            element.TryGetProperty(property, out var value) &&
            value.ValueKind == JsonValueKind.String)
        {
            return value.GetString() ?? fallback;
        }

        return fallback;
    }

    private static string FormatTimestamp(string timestamp)
    {
        if (DateTimeOffset.TryParse(timestamp, out var parsed))
        {
            return $"[{parsed.ToLocalTime():HH:mm:ss}] ";
        }

        return "";
    }

    private static void TrimCollection<T>(ObservableCollection<T> collection, int maxCount)
    {
        while (collection.Count > maxCount)
        {
            collection.RemoveAt(0);
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
    public bool IsLocal { get; set; }
    public bool IsProvisionalCpu { get; set; }
    public string? RuntimeStatus { get; set; }
    public bool CanVote { get; set; }
    public bool CanReplace { get; set; }
    public string StatusText => IsProvisionalCpu ? "CPU" : (IsHost ? "HOST" : (IsLocal ? "VOCE" : (!string.IsNullOrWhiteSpace(RuntimeStatus) ? RuntimeStatus!.ToUpperInvariant() : (IsConnected ? "ONLINE" : "OFFLINE"))));
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
