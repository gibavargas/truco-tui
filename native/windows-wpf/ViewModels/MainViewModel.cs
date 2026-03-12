using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using System.Windows;
using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using TrucoWPF.Models;
using TrucoWPF.Services;

namespace TrucoWPF.ViewModels;

public partial class MainViewModel : ObservableObject, IDisposable
{
    private readonly TrucoCoreService _core;
    private readonly IStringProvider _stringProvider;
    private CancellationTokenSource? _pollCts;
    private bool _disposed;

    [ObservableProperty]
    private string _status = StringProviderKeys.StatusWaiting;

    [ObservableProperty]
    private string _setupPlayerName = "Voce";

    [ObservableProperty]
    private int _setupNumPlayers = 1;

    [ObservableProperty]
    private int _setupLocaleIndex;

    [ObservableProperty]
    private string _mode = "idle";

    [ObservableProperty]
    private GameSnapshot? _snapshot;

    [ObservableProperty]
    private LobbySnapshot? _lobbySnapshot;

    [ObservableProperty]
    private string _inviteKeyInput = string.Empty;

    [ObservableProperty]
    private string _chatMessage = string.Empty;

    public System.Collections.ObjectModel.ObservableCollection<string> ChatEvents { get; } = new();
    public System.Collections.ObjectModel.ObservableCollection<LobbySlotItem> LobbySlots { get; } = new();

    public bool IsPlaying => Snapshot?.MatchFinished != true && Snapshot?.Players != null;
    public bool IsNotPlaying => !IsPlaying && Mode != "host_lobby" && Mode != "client_lobby";
    public bool IsOnlineLobby => Mode == "host_lobby" || Mode == "client_lobby";
    public bool IsMyTurn => Snapshot?.CurrentPlayerIdx == 0;

    public Visibility VisibilityIfNotPlaying => IsNotPlaying ? Visibility.Visible : Visibility.Collapsed;
    public Visibility VisibilityIfOnlineLobby => IsOnlineLobby ? Visibility.Visible : Visibility.Collapsed;
    public Visibility VisibilityIfHost => Mode == "host_lobby" ? Visibility.Visible : Visibility.Collapsed;
    public Visibility VisibilityIfInviteKey => string.IsNullOrEmpty(LobbySnapshot?.InviteKey) ? Visibility.Collapsed : Visibility.Visible;

    public Player? Me => Snapshot?.Players?.FirstOrDefault(p => p.ID == Snapshot.CurrentPlayerIdx);
    public Player? TopPlayer => GetPlayerAt(2);
    public Player? RightPlayer => Snapshot?.NumPlayers == 2 ? null : GetPlayerAt(1);
    public Player? LeftPlayer => Snapshot?.NumPlayers == 2 ? null : GetPlayerAt(3);

    public int UsPoints => GetTeamPoints(0);
    public int ThemPoints => GetTeamPoints(1);

    public bool ShowTrucoActions => Snapshot?.PendingRaiseFor != null && Snapshot.PendingRaiseFor == 0;
    public bool ShowAskTruco => Snapshot?.CanAskTruco == true && IsMyTurn;

    public string AskTrucoLabel => Snapshot?.PendingRaiseTo switch
    {
        3 => _stringProvider.Get(StringProviderKeys.TrucoLabel),
        6 => _stringProvider.Get(StringProviderKeys.SeisLabel),
        9 => _stringProvider.Get(StringProviderKeys.NoveLabel),
        12 => _stringProvider.Get(StringProviderKeys.DozeLabel),
        _ => "TRUCO!"
    };

    public string TrucoLabel => Snapshot?.CurrentHand?.Stake switch
    {
        1 => _stringProvider.Get(StringProviderKeys.TrucoLabel),
        3 => _stringProvider.Get(StringProviderKeys.SeisLabel),
        6 => _stringProvider.Get(StringProviderKeys.NoveLabel),
        9 => _stringProvider.Get(StringProviderKeys.DozeLabel),
        _ => ""
    };

    public int MyTeamID => 0;

    public string RoundText => Snapshot?.CurrentHand?.Round != null 
        ? _stringProvider.Format(StringProviderKeys.RoundFormat, Snapshot.CurrentHand.Round) 
        : "";

    public Visibility LeftPlayerVisibility => LeftPlayer != null ? Visibility.Visible : Visibility.Collapsed;

    public MainViewModel()
    {
        _core = new TrucoCoreService();
        _stringProvider = new StringProvider();
        _ = StartPollingAsync();
    }

    private Player? GetPlayerAt(int idx)
    {
        if (Snapshot?.Players == null || Snapshot.NumPlayers == null) return null;
        return Snapshot.NumPlayers == 2
            ? Snapshot.Players.FirstOrDefault(p => p.ID == (idx % 2))
            : Snapshot.Players.FirstOrDefault(p => p.ID == idx);
    }

    private int GetTeamPoints(int team)
    {
        if (Snapshot?.MatchPoints == null) return 0;
        return team == 0 
            ? (Snapshot.MatchPoints.TryGetValue("0", out var us) ? us : 0)
            : (Snapshot.MatchPoints.TryGetValue("1", out var them) ? them : 0);
    }

    private async Task StartPollingAsync()
    {
        _pollCts = new CancellationTokenSource();
        while (!_pollCts.Token.IsCancellationRequested)
        {
            try
            {
                var snapshotJson = _core.SnapshotJson();
                if (!string.IsNullOrEmpty(snapshotJson))
                {
                    await Application.Current.Dispatcher.InvokeAsync(() => ProcessSnapshot(snapshotJson));
                }
                
                var eventJson = _core.PollEventJson();
                if (!string.IsNullOrEmpty(eventJson))
                {
                    await Application.Current.Dispatcher.InvokeAsync(() => ProcessEvent(eventJson));
                }
                
                await Task.Delay(50, _pollCts.Token);
            }
            catch (OperationCanceledException)
            {
                break;
            }
            catch (Exception ex)
            {
                Debug.WriteLine($"Poll error: {ex.Message}");
            }
        }
    }

    private void ProcessSnapshot(string json)
    {
        try
        {
            var bundle = JsonSerializer.Deserialize<SnapshotBundle>(json);
            if (bundle != null)
            {
                Snapshot = bundle.Match;
                LobbySnapshot = bundle.Lobby;
                Mode = bundle.Mode ?? "idle";
                
                RebuildLobbySlots();

                OnPropertyChanged(nameof(IsPlaying));
                OnPropertyChanged(nameof(IsNotPlaying));
                OnPropertyChanged(nameof(IsOnlineLobby));
                OnPropertyChanged(nameof(VisibilityIfNotPlaying));
                OnPropertyChanged(nameof(VisibilityIfOnlineLobby));
                OnPropertyChanged(nameof(VisibilityIfHost));
                OnPropertyChanged(nameof(VisibilityIfInviteKey));
                OnPropertyChanged(nameof(IsMyTurn));
                OnPropertyChanged(nameof(Me));
                OnPropertyChanged(nameof(TopPlayer));
                OnPropertyChanged(nameof(RightPlayer));
                OnPropertyChanged(nameof(LeftPlayer));
                OnPropertyChanged(nameof(UsPoints));
                OnPropertyChanged(nameof(ThemPoints));
                OnPropertyChanged(nameof(ShowTrucoActions));
                OnPropertyChanged(nameof(ShowAskTruco));
                OnPropertyChanged(nameof(AskTrucoLabel));
                OnPropertyChanged(nameof(TrucoLabel));
                OnPropertyChanged(nameof(RoundText));
                OnPropertyChanged(nameof(LeftPlayerVisibility));
            }
        }
        catch (JsonException ex)
        {
            Debug.WriteLine($"Failed to parse snapshot: {ex.Message}");
        }
    }

    private void RebuildLobbySlots()
    {
        LobbySlots.Clear();
        if (LobbySnapshot == null || LobbySnapshot.Slots == null || LobbySnapshot.NumPlayers == null)
            return;

        int size = LobbySnapshot.Slots.Count > LobbySnapshot.NumPlayers.Value 
            ? LobbySnapshot.Slots.Count 
            : LobbySnapshot.NumPlayers.Value;

        var hostSeat = LobbySnapshot.HostSeat ?? 0;
        var connected = LobbySnapshot.ConnectedSeats ?? new List<int>();

        for (int i = 0; i < size; i++)
        {
            string label = i < LobbySnapshot.Slots.Count ? LobbySnapshot.Slots[i] : "";
            bool isAssigned = !string.IsNullOrEmpty(label);
            bool isHost = isAssigned && (i == hostSeat);
            bool isConnected = connected.Contains(i);

            LobbySlots.Add(new LobbySlotItem
            {
                Seat = i + 1,
                Label = isAssigned ? $"Slot {i + 1}: {label}" : $"Slot {i + 1}: (vaio)",
                IsAssigned = isAssigned,
                IsHost = isHost,
                IsConnected = isConnected,
                CanVote = isAssigned && !isHost,
                CanReplace = isAssigned && !isConnected
            });
        }
    }

    private void ProcessEvent(string json)
    {
        try
        {
            var appEvent = JsonSerializer.Deserialize<AppEvent>(json);
            if (appEvent != null && appEvent.Kind == "chat")
            {
                var payloadStr = appEvent.Payload?.ToString();
                if (!string.IsNullOrEmpty(payloadStr))
                {
                    ChatEvents.Add(payloadStr);
                }
            }
        }
        catch (JsonException ex)
        {
            Debug.WriteLine($"Event parse error: {ex.Message}");
        }
    }

    [RelayCommand]
    private void StartGame()
    {
        var localeIntent = new
        {
            kind = "set_locale",
            payload = new { locale = SetupLocaleIndex == 0 ? "pt-BR" : "en-US" }
        };
        _core.Dispatch(JsonSerializer.Serialize(localeIntent));

        var numPlayers = SetupNumPlayers == 1 ? 4 : 2;
        var names = numPlayers == 4
            ? new[] { string.IsNullOrEmpty(SetupPlayerName) ? "Voce" : SetupPlayerName, "CPU-Direita", "CPU-Parceiro", "CPU-Esquerda" }
            : new[] { string.IsNullOrEmpty(SetupPlayerName) ? "Voce" : SetupPlayerName, "CPU-Oponente" };
        var cpuFlags = numPlayers == 4
            ? new[] { false, true, true, true }
            : new[] { false, true };
        var gameIntent = new
        {
            kind = "new_offline_game",
            payload = new { player_names = names, cpu_flags = cpuFlags }
        };
        _core.Dispatch(JsonSerializer.Serialize(gameIntent));
        
        Status = _stringProvider.Format(StringProviderKeys.StatusPlaying, numPlayers);
    }

    [RelayCommand]
    private void PlayCard(Card? card)
    {
        if (card == null || Me?.Hand == null) return;

        var intent = new
        {
            kind = "game_action",
            payload = new { action = "play", card_index = Me.Hand.FindIndex(c => c.Rank == card.Rank && c.Suit == card.Suit) }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void RequestTruco()
    {
        var intent = new
        {
            kind = "game_action",
            payload = new { action = "truco" }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void AcceptTruco()
    {
        var intent = new
        {
            kind = "game_action",
            payload = new { action = "accept" }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void RefuseTruco()
    {
        var intent = new
        {
            kind = "game_action",
            payload = new { action = "refuse" }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void BackToSetup()
    {
        Snapshot = null;
        Mode = "idle";
        OnPropertyChanged(nameof(IsPlaying));
        OnPropertyChanged(nameof(IsNotPlaying));
        OnPropertyChanged(nameof(IsOnlineLobby));
        OnPropertyChanged(nameof(VisibilityIfNotPlaying));
        OnPropertyChanged(nameof(VisibilityIfOnlineLobby));
    }

    [RelayCommand]
    private void HostOnlineGame()
    {
        var numPlayers = SetupNumPlayers == 1 ? 4 : 2;
        var name = string.IsNullOrEmpty(SetupPlayerName) ? "Host" : SetupPlayerName;
        var intent = new
        {
            kind = "create_host_session",
            payload = new { host_name = name, num_players = numPlayers }
        };
        ChatEvents.Clear();
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void JoinOnlineGame()
    {
        if (string.IsNullOrWhiteSpace(InviteKeyInput)) return;
        var name = string.IsNullOrEmpty(SetupPlayerName) ? "Client" : SetupPlayerName;
        var intent = new
        {
            kind = "join_session",
            payload = new { client_name = name, invite_key = InviteKeyInput.Trim(), preferred_role = "auto" }
        };
        ChatEvents.Clear();
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void StartHostedMatch()
    {
        var intent = new { kind = "start_hosted_match" };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void SendChat()
    {
        if (string.IsNullOrWhiteSpace(ChatMessage)) return;
        var intent = new
        {
            kind = "send_chat",
            payload = new { text = ChatMessage }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
        ChatMessage = string.Empty;
    }

    [RelayCommand]
    private void VoteHost(LobbySlotItem? slot)
    {
        if (slot == null) return;
        var intent = new
        {
            kind = "vote_host",
            payload = new { seat_index = slot.Seat - 1 }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void RequestReplacementInvite(LobbySlotItem? slot)
    {
        if (slot == null) return;
        var intent = new
        {
            kind = "request_replacement_invite",
            payload = new { target_seat_index = slot.Seat - 1 }
        };
        _core.Dispatch(JsonSerializer.Serialize(intent));
    }

    [RelayCommand]
    private void LeaveOnlineGame()
    {
        var intent = new { kind = "close_session" };
        _core.Dispatch(JsonSerializer.Serialize(intent));
        BackToSetup();
    }

    public void Dispose()
    {
        if (_disposed) return;
        _disposed = true;
        _pollCts?.Cancel();
        _pollCts?.Dispose();
        _core.Dispose();
    }
}
