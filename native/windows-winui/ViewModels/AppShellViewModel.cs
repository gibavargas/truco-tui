using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using Microsoft.UI.Dispatching;
using Microsoft.UI.Xaml;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Globalization;
using System.Linq;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Windows.ApplicationModel.DataTransfer;
using TrucoWinUI.Models;
using TrucoWinUI.Services;

namespace TrucoWinUI.ViewModels;

public partial class AppShellViewModel : ObservableObject, IDisposable
{
    private readonly TrucoCoreService _core;
    private readonly DispatcherQueue _dispatcherQueue;
    private readonly CancellationTokenSource _cts = new();

    private SnapshotBundle _bundle = new();
    private string _menuPane = "home";

    public ObservableCollection<LobbySeatViewModel> LobbySeats { get; } = [];
    public ObservableCollection<LobbySeatViewModel> CandidateSeats { get; } = [];
    public ObservableCollection<LobbySeatViewModel> ReplacementSeats { get; } = [];
    public ObservableCollection<ActivityEntry> ChatFeed { get; } = [];
    public ObservableCollection<string> MatchLog { get; } = [];
    public ObservableCollection<string> DiagnosticsLog { get; } = [];

    [ObservableProperty]
    private string statusText = "Runtime inicializando";

    [ObservableProperty]
    private string infoBannerText = string.Empty;

    [ObservableProperty]
    private string errorBannerText = string.Empty;

    [ObservableProperty]
    private string currentModeText = "idle";

    [ObservableProperty]
    private string versionText = string.Empty;

    [ObservableProperty]
    private string localeCode = "pt-BR";

    [ObservableProperty]
    private string localeDisplayText = "Idioma: Português (BR)";

    [ObservableProperty]
    private string inviteKey = string.Empty;

    [ObservableProperty]
    private string replacementInviteKey = string.Empty;

    [ObservableProperty]
    private string connectionDetails = string.Empty;

    [ObservableProperty]
    private string handSummary = string.Empty;

    [ObservableProperty]
    private string pendingRaiseText = string.Empty;

    [ObservableProperty]
    private string offlinePlayerName = Environment.UserName;

    [ObservableProperty]
    private int offlineNumPlayers = 2;

    [ObservableProperty]
    private string hostName = Environment.UserName;

    [ObservableProperty]
    private int hostNumPlayers = 2;

    [ObservableProperty]
    private string bindAddress = string.Empty;

    [ObservableProperty]
    private string relayUrl = string.Empty;

    [ObservableProperty]
    private string transportMode = "tcp_tls";

    [ObservableProperty]
    private string joinKey = string.Empty;

    [ObservableProperty]
    private string joinPlayerName = Environment.UserName;

    [ObservableProperty]
    private string desiredRole = "auto";

    [ObservableProperty]
    private string chatInput = string.Empty;

    [ObservableProperty]
    private LobbySeatViewModel? selectedCandidateSeat;

    [ObservableProperty]
    private LobbySeatViewModel? selectedReplacementSeat;

    [ObservableProperty]
    private TableSeatViewModel bottomSeat = new();

    [ObservableProperty]
    private TableSeatViewModel topSeat = new();

    [ObservableProperty]
    private TableSeatViewModel leftSeat = new();

    [ObservableProperty]
    private TableSeatViewModel rightSeat = new();

    [ObservableProperty]
    private Visibility homeVisibility = Visibility.Visible;

    [ObservableProperty]
    private Visibility offlineVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility hostVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility joinVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility lobbyVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility gameVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility inviteVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility replacementInviteVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility leftSeatVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility rightSeatVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private Visibility topSeatVisibility = Visibility.Collapsed;

    [ObservableProperty]
    private string localSeatTitle = "Você";

    [ObservableProperty]
    private bool canStartHostedMatch;

    [ObservableProperty]
    private bool canStartNewHand;

    [ObservableProperty]
    private bool canRequestTruco;

    [ObservableProperty]
    private bool canAnswerRaise;

    [ObservableProperty]
    private bool canPlayCards;

    [ObservableProperty]
    private bool canSendChat;

    [ObservableProperty]
    private bool hasActiveSession;

    [ObservableProperty]
    private bool canResetSession;

    [ObservableProperty]
    private bool isLobbyScreen;

    [ObservableProperty]
    private bool isGameScreen;

    [ObservableProperty]
    private bool isMenuScreen = true;

    [ObservableProperty]
    private bool isDiagnosticsOpen;

    public AppShellViewModel()
    {
        _dispatcherQueue = DispatcherQueue.GetForCurrentThread();
        _core = new TrucoCoreService();
        VersionText = BuildVersionText(_core.GetVersions());
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: true);
        _ = PollLoopAsync(_cts.Token);
    }

    public void Dispose()
    {
        _cts.Cancel();
        _core.CloseSession();
        _core.Dispose();
        _cts.Dispose();
    }

    [RelayCommand]
    private void ShowHome() => SetMenuPane("home");

    [RelayCommand]
    private void ShowOfflineSetup() => SetMenuPane("offline");

    [RelayCommand]
    private void ShowHostSetup() => SetMenuPane("host");

    [RelayCommand]
    private void ShowJoinSetup() => SetMenuPane("join");

    [RelayCommand]
    private void ToggleDiagnostics() => IsDiagnosticsOpen = !IsDiagnosticsOpen;

    [RelayCommand]
    private void OpenDiagnostics() => IsDiagnosticsOpen = true;

    [RelayCommand]
    private void CloseDiagnostics() => IsDiagnosticsOpen = false;

    [RelayCommand]
    private void CopyInviteKey() => CopyTextToClipboard(InviteKey);

    [RelayCommand]
    private void CopyReplacementInvite() => CopyTextToClipboard(ReplacementInviteKey);

    [RelayCommand]
    private void ToggleLocale()
    {
        string nextLocale = LocaleCode == "en-US" ? "pt-BR" : "en-US";
        AppError? error = _core.SetLocale(nextLocale);
        HandleDispatchResult(error, nextLocale == "en-US"
            ? "Idioma alterado para English (US)."
            : "Idioma alterado para Português (BR).");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: true);
    }

    [RelayCommand]
    private void StartOfflineGame()
    {
        string playerName = NormalizeName(OfflinePlayerName);
        List<string> names = [playerName, "CPU-2"];
        List<bool> cpu = [false, true];
        if (OfflineNumPlayers == 4)
        {
            names.Add("CPU-3");
            names.Add("CPU-4");
            cpu.Add(true);
            cpu.Add(true);
        }

        AppError? error = _core.StartOfflineGame(names.ToArray(), cpu.ToArray());
        HandleDispatchResult(error, "Partida offline criada.");
        if (error is null)
        {
            ChatFeed.Clear();
        }
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void NewHand()
    {
        HandleDispatchResult(_core.NewHand(), "Nova mão iniciada.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void CreateHostSession()
    {
        AppError? error = _core.CreateHostSession(
            NormalizeName(HostName),
            HostNumPlayers,
            NullIfWhitespace(BindAddress),
            NullIfWhitespace(RelayUrl),
            NullIfWhitespace(TransportMode));
        HandleDispatchResult(error, "Sessão host criada.");
        if (error is null)
        {
            ChatFeed.Clear();
        }
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void JoinOnlineSession()
    {
        AppError? error = _core.JoinSession(
            JoinKey.Trim(),
            NormalizeName(JoinPlayerName),
            string.IsNullOrWhiteSpace(DesiredRole) ? "auto" : DesiredRole.Trim());
        HandleDispatchResult(error, "Sessão conectada.");
        if (error is null)
        {
            ChatFeed.Clear();
        }
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void StartHostedMatch()
    {
        HandleDispatchResult(_core.StartHostedMatch(), "Partida online iniciada.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void CloseSession()
    {
        HandleDispatchResult(_core.CloseSession(), "Sessão encerrada.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void ResetSession()
    {
        HandleDispatchResult(_core.ResetSession(), "Sessão reiniciada.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void SendChat()
    {
        string text = ChatInput.Trim();
        if (string.IsNullOrWhiteSpace(text))
        {
            return;
        }

        HandleDispatchResult(_core.SendChat(text), "Mensagem enviada.");
        ChatInput = string.Empty;
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void VoteHost()
    {
        if (SelectedCandidateSeat is null)
        {
            ErrorBannerText = "Selecione um slot para votar host.";
            return;
        }

        HandleDispatchResult(_core.VoteHost(SelectedCandidateSeat.SeatIndex), "Voto de host enviado.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void RequestReplacementInvite()
    {
        if (SelectedReplacementSeat is null)
        {
            ErrorBannerText = "Selecione um slot para gerar convite de reposição.";
            return;
        }

        HandleDispatchResult(_core.RequestReplacementInvite(SelectedReplacementSeat.SeatIndex), "Pedido de substituição enviado.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void PlayCard(CardState? card)
    {
        if (card is null || BottomSeat.HandCards.Count == 0)
        {
            return;
        }

        int index = BottomSeat.HandCards.FindIndex(c => c.Card is not null && ReferenceEquals(c.Card, card));
        if (index < 0)
        {
            index = BottomSeat.HandCards.FindIndex(c => c.Card is not null && c.Card.Rank == card.Rank && c.Card.Suit == card.Suit);
        }

        if (index < 0)
        {
            ErrorBannerText = "Carta não encontrada na mão local.";
            return;
        }

        HandleDispatchResult(_core.PlayCard(index), "Carta enviada.");
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    [RelayCommand]
    private void RequestTruco() => DispatchGameAction(_core.RequestTruco(), "Pedido de truco enviado.");

    [RelayCommand]
    private void AcceptTruco() => DispatchGameAction(_core.AcceptTruco(), "Truco aceito.");

    [RelayCommand]
    private void RefuseTruco() => DispatchGameAction(_core.RefuseTruco(), "Truco recusado.");

    private void DispatchGameAction(AppError? error, string successMessage)
    {
        HandleDispatchResult(error, successMessage);
        RefreshSnapshot(_core.GetSnapshot(), preserveMenuPane: false);
    }

    private async Task PollLoopAsync(CancellationToken token)
    {
        while (!token.IsCancellationRequested)
        {
            await Task.Delay(120, token);
            SnapshotBundle bundle = _core.GetSnapshot();
            List<AppEvent> drained = [];
            AppEvent? appEvent;
            while ((appEvent = _core.PollEvent()) is not null)
            {
                drained.Add(appEvent);
            }

            _dispatcherQueue.TryEnqueue(() =>
            {
                foreach (AppEvent ev in drained)
                {
                    HandleEvent(ev);
                }
                RefreshSnapshot(bundle, preserveMenuPane: false);
            });
        }
    }

    private void RefreshSnapshot(SnapshotBundle bundle, bool preserveMenuPane)
    {
        NormalizeBundle(bundle);
        _bundle = bundle;
        if (bundle.Mode == "idle" && !preserveMenuPane)
        {
            ChatFeed.Clear();
            InfoBannerText = string.Empty;
        }
        LocaleCode = bundle.Locale;
        LocaleDisplayText = BuildLocaleDisplayText(bundle.Locale);
        CurrentModeText = bundle.Mode;
        InviteKey = bundle.Lobby?.InviteKey ?? string.Empty;
        if (bundle.Mode == "idle")
        {
            ReplacementInviteKey = string.Empty;
        }
        ConnectionDetails = BuildConnectionDetails(bundle);
        StatusText = BuildStatusText(bundle);
        ErrorBannerText = bundle.Connection.LastError?.Message ?? string.Empty;

        MatchLog.ReplaceWith(bundle.Match?.Logs ?? []);
        DiagnosticsLog.ReplaceWith(bundle.Diagnostics.EventLog ?? []);
        UpdateScreenState(bundle.Mode, preserveMenuPane);
        UpdateLobby(bundle.Lobby);
        UpdateTable(bundle.Match);
        UpdateActionState(bundle);
        OnPropertyChanged(nameof(UsScore));
        OnPropertyChanged(nameof(ThemScore));
        OnPropertyChanged(nameof(StakeText));
        OnPropertyChanged(nameof(ViraText));
        OnPropertyChanged(nameof(ManilhaText));
        OnPropertyChanged(nameof(ViraCardViewModel));
        OnPropertyChanged(nameof(ManilhaCardViewModel));
        OnPropertyChanged(nameof(CurrentTurnText));
        OnPropertyChanged(nameof(WinnerText));
    }

    public int UsScore => _bundle.Match?.MatchPoints.GetValueOrDefault(0) ?? 0;
    public int ThemScore => _bundle.Match?.MatchPoints.GetValueOrDefault(1) ?? 0;
    public string StakeText => (_bundle.Match?.CurrentHand.Stake ?? 1).ToString(CultureInfo.InvariantCulture);
    public string ViraText => _bundle.Match?.CurrentHand?.Vira?.ShortLabel ?? "--";
    public string ManilhaText => _bundle.Match?.CurrentHand?.Manilha ?? "--";

    public HandCardViewModel? ViraCardViewModel => _bundle.Match?.CurrentHand?.Vira is { Rank: not "" } vira
        ? new HandCardViewModel { Card = vira, IsFaceUp = true, Scale = 1.0, Rotation = 0 }
        : null;

    public HandCardViewModel? ManilhaCardViewModel => !string.IsNullOrWhiteSpace(_bundle.Match?.CurrentHand?.Manilha)
        ? new HandCardViewModel { Card = new CardState { Rank = _bundle.Match.CurrentHand.Manilha, Suit = "" }, IsFaceUp = true, Scale = 1.0, Rotation = 0 }
        : null;

    public string CurrentTurnText => BuildCurrentTurnText(_bundle.Match, BottomSeat);
    public string WinnerText => _bundle.Match?.MatchFinished == true ? $"Time vencedor: {_bundle.Match.WinnerTeam}" : string.Empty;

    private void UpdateScreenState(string mode, bool preserveMenuPane)
    {
        if (!preserveMenuPane && mode.EndsWith("_lobby", StringComparison.Ordinal))
        {
            _menuPane = "lobby";
        }
        else if (!preserveMenuPane && mode.EndsWith("_match", StringComparison.Ordinal))
        {
            _menuPane = "game";
        }
        else if (!preserveMenuPane && mode == "idle" && (_menuPane == "lobby" || _menuPane == "game"))
        {
            _menuPane = "home";
        }

        HomeVisibility = _menuPane == "home" && mode == "idle" ? Visibility.Visible : Visibility.Collapsed;
        OfflineVisibility = _menuPane == "offline" && mode == "idle" ? Visibility.Visible : Visibility.Collapsed;
        HostVisibility = _menuPane == "host" && mode == "idle" ? Visibility.Visible : Visibility.Collapsed;
        JoinVisibility = _menuPane == "join" && mode == "idle" ? Visibility.Visible : Visibility.Collapsed;
        LobbyVisibility = mode.EndsWith("_lobby", StringComparison.Ordinal) ? Visibility.Visible : Visibility.Collapsed;
        GameVisibility = mode.EndsWith("_match", StringComparison.Ordinal) ? Visibility.Visible : Visibility.Collapsed;
        InviteVisibility = string.IsNullOrWhiteSpace(InviteKey) ? Visibility.Collapsed : Visibility.Visible;
        ReplacementInviteVisibility = string.IsNullOrWhiteSpace(ReplacementInviteKey) ? Visibility.Collapsed : Visibility.Visible;
        IsLobbyScreen = LobbyVisibility == Visibility.Visible;
        IsGameScreen = GameVisibility == Visibility.Visible;
        IsMenuScreen = !IsLobbyScreen && !IsGameScreen;
    }

    private void UpdateLobby(LobbySnapshot? lobby)
    {
        List<LobbySeatViewModel> seats = [];
        Dictionary<int, int> seatProtocolVersions = _bundle.Connection.Network?.SeatProtocolVersions ?? [];
        if (lobby is not null)
        {
            for (int i = 0; i < lobby.NumPlayers; i++)
            {
                string name = i < lobby.Slots.Count ? lobby.Slots[i] : string.Empty;
                bool connected = lobby.ConnectedSeats.GetValueOrDefault(i) || i == 0 && (lobby.HostSeat == 0 || lobby.AssignedSeat == 0);
                seats.Add(new LobbySeatViewModel
                {
                    SeatIndex = i,
                    Name = string.IsNullOrWhiteSpace(name) ? "assento vazio" : name,
                    IsAssigned = i == lobby.AssignedSeat,
                    IsConnected = connected,
                    IsHost = i == lobby.HostSeat,
                    IsEmpty = string.IsNullOrWhiteSpace(name),
                    ProtocolVersion = seatProtocolVersions.GetValueOrDefault(i),
                    StatusText = connected ? "conectado" : lobby.Started && !string.IsNullOrWhiteSpace(name) ? "aguardando reconexão" : "livre",
                });
            }
        }

        LobbySeats.ReplaceWith(seats);
        CandidateSeats.ReplaceWith(seats.Where(s => !s.IsEmpty));
        ReplacementSeats.ReplaceWith(seats.Where(s => s.SeatIndex > 0 && !s.IsEmpty));
        SelectedCandidateSeat = CandidateSeats.FirstOrDefault(s => s.SeatIndex == SelectedCandidateSeat?.SeatIndex) ?? CandidateSeats.FirstOrDefault();
        SelectedReplacementSeat = ReplacementSeats.FirstOrDefault(s => s.SeatIndex == SelectedReplacementSeat?.SeatIndex) ?? ReplacementSeats.FirstOrDefault();
        CanStartHostedMatch = _bundle.Mode == "host_lobby" && lobby is not null && lobby.Slots.Count == lobby.NumPlayers && lobby.Slots.All(s => !string.IsNullOrWhiteSpace(s));
    }

    private void UpdateTable(MatchSnapshot? match)
    {
        if (match is null || match.Players.Count == 0)
        {
            BottomSeat = new TableSeatViewModel();
            TopSeat = new TableSeatViewModel();
            LeftSeat = new TableSeatViewModel();
            RightSeat = new TableSeatViewModel();
            LeftSeatVisibility = Visibility.Collapsed;
            RightSeatVisibility = Visibility.Collapsed;
            TopSeatVisibility = Visibility.Collapsed;
            HandSummary = string.Empty;
            PendingRaiseText = string.Empty;
            return;
        }

        int localIndex = match.CurrentPlayerIdx >= 0 ? match.CurrentPlayerIdx : 0;
        List<TableSeatViewModel> layout = BuildTableLayout(match, localIndex);
        BottomSeat = layout[0];
        TopSeat = layout[1];
        LeftSeat = layout[2];
        RightSeat = layout[3];
        LocalSeatTitle = BottomSeat.Name;
        TopSeatVisibility = TopSeat.IsVisible ? Visibility.Visible : Visibility.Collapsed;
        LeftSeatVisibility = LeftSeat.IsVisible ? Visibility.Visible : Visibility.Collapsed;
        RightSeatVisibility = RightSeat.IsVisible ? Visibility.Visible : Visibility.Collapsed;
        HandSummary = $"Mão local: {BottomSeat.HandCount} cartas  |  Rodada {match.CurrentHand.Round + 1}";
        PendingRaiseText = BuildRaiseSummary(match);
    }

    private void UpdateActionState(SnapshotBundle bundle)
    {
        MatchSnapshot? match = bundle.Match;
        bool hasMatch = match is not null;
        HasActiveSession = bundle.Mode != "idle";
        CanResetSession = HasActiveSession;
        CanSendChat = HasActiveSession;
        CanStartNewHand = hasMatch && (bundle.Mode == "offline_match" || bundle.Mode == "host_match");
        CanRequestTruco = hasMatch &&
            BottomSeat.IsCurrentTurn &&
            (match!.CanAskTruco || match.PendingRaiseFor == BottomSeat.TeamIndex);
        CanAnswerRaise = hasMatch && match!.PendingRaiseFor != -1 && BottomSeat.TeamIndex == match.PendingRaiseFor;
        CanPlayCards = hasMatch && BottomSeat.IsCurrentTurn && match!.PendingRaiseFor == -1 && BottomSeat.HandCount > 0;
    }

    private void HandleEvent(AppEvent appEvent)
    {
        switch (appEvent.Kind)
        {
            case "error":
                AppError? error = DeserializePayload<AppError>(appEvent.Payload);
                if (error is not null)
                {
                    ErrorBannerText = error.Message;
                    AddChatEntry("error", error.Message, "#FF7070", appEvent.Timestamp);
                }
                break;
            case "locale_changed":
                if (appEvent.Payload.TryGetProperty("locale", out JsonElement localeEl))
                {
                    string locale = localeEl.GetString() ?? LocaleCode;
                    LocaleCode = locale;
                    LocaleDisplayText = BuildLocaleDisplayText(locale);
                    AddChatEntry("system", $"Idioma alterado para {BuildLocaleName(locale)}.", "#80C8FF", appEvent.Timestamp);
                }
                break;
            case "chat":
                string author = appEvent.Payload.TryGetProperty("author", out JsonElement authorEl) ? authorEl.GetString() ?? "chat" : "chat";
                string text = appEvent.Payload.TryGetProperty("text", out JsonElement textEl) ? textEl.GetString() ?? string.Empty : string.Empty;
                AddChatEntry("chat", $"{author}: {text}", "#7EE787", appEvent.Timestamp);
                break;
            case "system":
                string systemText = appEvent.Payload.TryGetProperty("text", out JsonElement systemEl) ? systemEl.GetString() ?? string.Empty : appEvent.Payload.ToString();
                InfoBannerText = systemText;
                AddChatEntry("system", systemText, "#80C8FF", appEvent.Timestamp);
                break;
            case "host_created":
                if (appEvent.Payload.TryGetProperty("invite_key", out JsonElement inviteEl))
                {
                    InviteKey = inviteEl.GetString() ?? string.Empty;
                    AddChatEntry("system", "Convite do host atualizado.", "#80C8FF", appEvent.Timestamp);
                }
                break;
            case "replacement_invite":
                if (appEvent.Payload.TryGetProperty("invite_key", out JsonElement replacementEl))
                {
                    ReplacementInviteKey = replacementEl.GetString() ?? string.Empty;
                    AddChatEntry("system", "Convite de reposição gerado.", "#FFD166", appEvent.Timestamp);
                }
                break;
            case "client_joined":
                AddChatEntry("system", "Jogador entrou na sessão.", "#80C8FF", appEvent.Timestamp);
                break;
            case "lobby_updated":
                AddChatEntry("system", "Lobby atualizado.", "#80C8FF", appEvent.Timestamp);
                break;
            case "match_updated":
                AddChatEntry("system", "Mesa atualizada.", "#80C8FF", appEvent.Timestamp);
                break;
            case "failover_promoted":
            case "failover_rejoined":
            case "session_ready":
            case "match_started":
            case "session_closed":
                AddChatEntry("system", BuildSystemEventMessage(appEvent.Kind), "#80C8FF", appEvent.Timestamp);
                break;
        }
    }

    private void HandleDispatchResult(AppError? error, string successMessage)
    {
        if (error is null)
        {
            ErrorBannerText = string.Empty;
            InfoBannerText = successMessage;
            return;
        }

        ErrorBannerText = error.Message;
    }

    private void SetMenuPane(string pane)
    {
        _menuPane = pane;
        RefreshSnapshot(_bundle, preserveMenuPane: true);
    }

    private static string BuildVersionText(CoreVersions versions)
        => $"Core API {versions.CoreApiVersion}  |  Protocolo {versions.ProtocolVersion}  |  Snapshot {versions.SnapshotSchemaVersion}";

    private static string BuildLocaleDisplayText(string locale)
        => $"Idioma: {BuildLocaleName(locale)}";

    private static string BuildLocaleName(string locale)
        => locale == "en-US" ? "English (US)" : "Português (BR)";

    private static string BuildSystemEventMessage(string kind)
        => kind switch
        {
            "session_ready" => "Sessão pronta.",
            "match_started" => "Partida iniciada.",
            "session_closed" => "Sessão encerrada.",
            "failover_promoted" => "Failover promovido.",
            "failover_rejoined" => "Reconectado após failover.",
            _ => kind.Replace('_', ' '),
        };

    private static string BuildStatusText(SnapshotBundle bundle)
    {
        if (bundle.Connection.LastError is not null)
        {
            return $"{bundle.Mode} com erro";
        }

        return bundle.Mode switch
        {
            "idle" => "Pronto",
            "offline_match" => "Partida offline em andamento",
            "host_lobby" => "Lobby do host",
            "host_match" => "Partida online como host",
            "client_lobby" => "Lobby conectado",
            "client_match" => "Partida online conectada",
            _ => bundle.Mode,
        };
    }

    private static string BuildConnectionDetails(SnapshotBundle bundle)
    {
        NetworkSnapshot? network = bundle.Connection.Network;
        string transport = FormatTransport(network?.Transport);
        string supported = FormatSupportedVersions(network);
        string compatibility = BuildCompatibilitySummary(bundle, network);
        return $"Status: {bundle.Connection.Status}  |  Rede: {transport}  |  Compat: {compatibility}  |  Build: {supported}";
    }

    private static string BuildCompatibilitySummary(SnapshotBundle bundle, NetworkSnapshot? network)
    {
        if (network is null)
        {
            return "offline";
        }

        if (bundle.Connection.IsHost)
        {
            IEnumerable<string> seatVersions = network.SeatProtocolVersions
                .Values
                .Where(version => version > 0)
                .Distinct()
                .OrderByDescending(version => version)
                .Select(version => $"v{version}");
            string joined = string.Join("/", seatVersions);
            if (string.IsNullOrWhiteSpace(joined))
            {
                joined = FormatSupportedVersions(network);
            }
            return network.MixedProtocolSession ? $"misto {joined}" : joined;
        }

        return network.NegotiatedProtocolVersion > 0
            ? $"negociado v{network.NegotiatedProtocolVersion}"
            : FormatSupportedVersions(network);
    }

    private static string FormatSupportedVersions(NetworkSnapshot? network)
    {
        if (network is null || network.SupportedProtocolVersions.Count == 0)
        {
            return "-";
        }
        return string.Join("/", network.SupportedProtocolVersions.Select(version => $"v{version}"));
    }

    private static string FormatTransport(string? transport)
        => transport == "relay_quic_v2" ? "Relay QUIC v2" : "TCP + TLS";

    private static string BuildRaiseSummary(MatchSnapshot match)
    {
        if (match.PendingRaiseFor == -1)
        {
            return string.Empty;
        }

        return $"Truco pendente: time {match.PendingRaiseFor} responde por {match.PendingRaiseTo}";
    }

    private static string BuildCurrentTurnText(MatchSnapshot? match, TableSeatViewModel bottomSeat)
    {
        if (match is null)
        {
            return string.Empty;
        }

        if (match.PendingRaiseFor != -1 && bottomSeat.TeamIndex == match.PendingRaiseFor)
        {
            return "Seu time responde ao truco";
        }

        return bottomSeat.IsCurrentTurn ? "Sua vez" : "Aguardando jogada";
    }

    private static List<TableSeatViewModel> BuildTableLayout(MatchSnapshot match, int localIndex)
    {
        Dictionary<int, PlayedCardState> playedByPlayerId = (match.CurrentHand?.RoundCards ?? [])
            .ToDictionary(card => card.PlayerId, card => card);

        int playerCount = match.Players.Count;
        TableSeatViewModel[] seats = [new(), new(), new(), new()];
        for (int relative = 0; relative < Math.Min(playerCount, 4); relative++)
        {
            int playerIndex = (localIndex + relative) % playerCount;
            PlayerState player = match.Players[playerIndex];
            TableSeatViewModel seat = ToSeatViewModel(playerIndex, relative, player, match, playedByPlayerId);
            switch (relative)
            {
                case 0:
                    seats[0] = seat;
                    break;
                case 1:
                    seats[playerCount == 2 ? 1 : 3] = seat;
                    break;
                case 2:
                    seats[1] = seat;
                    break;
                case 3:
                    seats[2] = seat;
                    break;
            }
        }

        return seats.ToList();
    }

    private static TableSeatViewModel ToSeatViewModel(int playerIndex, int relativeIndex, PlayerState player, MatchSnapshot match, IReadOnlyDictionary<int, PlayedCardState> playedByPlayerId)
    {
        string role = relativeIndex switch
        {
            0 => "Você",
            1 when match.NumPlayers == 2 => "Oponente",
            1 => "Direita",
            2 => "Oponente",
            3 => "Esquerda",
            _ => string.Empty,
        };

        return new TableSeatViewModel
        {
            SeatIndex = playerIndex,
            PlayerId = player.Id,
            Name = player.Name,
            RoleLabel = role,
            TeamIndex = player.Team,
            TeamLabel = $"Time {player.Team}",
            IsVisible = true,
            IsLocal = relativeIndex == 0,
            IsCurrentTurn = player.Id == match.TurnPlayer,
            IsCpu = player.Cpu,
            IsProvisionalCpu = player.ProvisionalCpu,
            HandCount = player.Hand.Count,
            HandCards = BuildHandVisuals(player.Hand, relativeIndex == 0),
            PlayedCard = playedByPlayerId.TryGetValue(player.Id, out PlayedCardState? played) ? played.Card : null,
            PlayedCardViewModel = playedByPlayerId.TryGetValue(player.Id, out PlayedCardState? playedCard)
                ? new HandCardViewModel
                {
                    Card = playedCard.FaceDown ? new CardState() : playedCard.Card,
                    IsFaceUp = !playedCard.FaceDown,
                    Scale = 0.85,
                    Rotation = 0
                }
                : null
        };
    }

    private static List<HandCardViewModel> BuildHandVisuals(IReadOnlyList<CardState> cards, bool isLocalSeat)
    {
        if (cards.Count == 0)
        {
            return [];
        }

        int visibleCount = isLocalSeat ? cards.Count : Math.Min(cards.Count, 3);
        double rotationStep = isLocalSeat ? 9.0 : 7.0;
        double scale = isLocalSeat ? 1.0 : 0.72;
        double midpoint = (visibleCount - 1) / 2.0;
        List<HandCardViewModel> result = [];

        for (int index = 0; index < visibleCount; index++)
        {
            double offset = index - midpoint;
            result.Add(new HandCardViewModel
            {
                Card = isLocalSeat ? cards[index] : null,
                IsFaceUp = isLocalSeat,
                Rotation = offset * rotationStep,
                Scale = scale,
            });
        }

        return result;
    }

    private static T? DeserializePayload<T>(JsonElement payload)
    {
        if (payload.ValueKind is JsonValueKind.Null or JsonValueKind.Undefined)
        {
            return default;
        }

        return JsonSerializer.Deserialize<T>(payload.GetRawText(), new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
    }

    private void AddChatEntry(string channel, string text, string accent, string timestamp)
    {
        if (string.IsNullOrWhiteSpace(text))
        {
            return;
        }

        ChatFeed.Add(new ActivityEntry
        {
            Channel = channel,
            Text = text,
            Accent = accent,
            Timestamp = string.IsNullOrWhiteSpace(timestamp) ? DateTime.Now.ToString("HH:mm:ss", CultureInfo.InvariantCulture) : timestamp,
        });

        while (ChatFeed.Count > 200)
        {
            ChatFeed.RemoveAt(0);
        }
    }

    private static void CopyTextToClipboard(string? text)
    {
        if (string.IsNullOrWhiteSpace(text))
        {
            return;
        }

        DataPackage package = new();
        package.SetText(text);
        Clipboard.SetContent(package);
    }

    private static string NormalizeName(string? value)
        => string.IsNullOrWhiteSpace(value) ? "Jogador" : value.Trim();

    private static string? NullIfWhitespace(string? value)
        => string.IsNullOrWhiteSpace(value) ? null : value.Trim();

    private static string FormatSeatIndex(int seatIndex)
        => seatIndex < 0 ? "-" : (seatIndex + 1).ToString(CultureInfo.InvariantCulture);

    private static void NormalizeBundle(SnapshotBundle bundle)
    {
        bundle.Lobby ??= new LobbySnapshot();
        bundle.Lobby.Slots ??= [];
        bundle.Lobby.ConnectedSeats ??= [];
        if (bundle.Connection.Network is not null)
        {
            bundle.Connection.Network.SupportedProtocolVersions ??= [];
            bundle.Connection.Network.SeatProtocolVersions ??= [];
        }

        if (bundle.Match is null)
        {
            bundle.Diagnostics.EventLog ??= [];
            return;
        }

        bundle.Match.Players ??= [];
        bundle.Match.Logs ??= [];
        bundle.Match.MatchPoints ??= [];
        bundle.Match.CurrentHand ??= new HandState();
        bundle.Match.CurrentHand.RoundCards ??= [];
        bundle.Match.CurrentHand.TrickResults ??= [];
        bundle.Match.CurrentHand.TrickWins ??= [];
        foreach (PlayerState player in bundle.Match.Players)
        {
            player.Hand ??= [];
        }

        bundle.Diagnostics.EventLog ??= [];
    }
}
