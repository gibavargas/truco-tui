using CommunityToolkit.Mvvm.ComponentModel;
using CommunityToolkit.Mvvm.Input;
using Microsoft.UI.Dispatching;
using System;
using System.Diagnostics;
using System.Text.Json;
using System.Threading.Tasks;
using TrucoWinUI.Models;
using TrucoWinUI.Services;

namespace TrucoWinUI.ViewModels;

public partial class AppShellViewModel : ObservableObject
{
    private readonly TrucoCoreService _core = new();
    private readonly DispatcherQueue _dispatcherQueue = DispatcherQueue.GetForCurrentThread();

    [ObservableProperty]
    private string status = "Runtime aguardando inicializacao";

    [ObservableProperty]
    [NotifyPropertyChangedFor(nameof(IsPlaying))]
    [NotifyPropertyChangedFor(nameof(IsNotPlaying))]
    [NotifyPropertyChangedFor(nameof(IsMyTurn))]
    private GameSnapshot? snapshot;

    public bool IsPlaying => Snapshot != null;
    public bool IsNotPlaying => Snapshot == null;
    public bool IsMyTurn => Snapshot?.TurnPlayer == 0;

    public AppShellViewModel()
    {
        _ = PollLoopAsync();
    }

    private async Task PollLoopAsync()
    {
        while (true)
        {
            await Task.Delay(50);

            var eventJson = _core.PollEventJson();
            if (eventJson != null)
            {
                RefreshSnapshot();
            }
        }
    }

    private void RefreshSnapshot()
    {
        var json = _core.SnapshotJson();
        if (json != null)
        {
            _dispatcherQueue.TryEnqueue(() =>
            {
                try
                {
                    Snapshot = JsonSerializer.Deserialize<GameSnapshot>(json, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                }
                catch (Exception ex)
                {
                    Debug.WriteLine($"Failed to parse snapshot: {ex.Message}");
                }
            });
        }
    }

    [RelayCommand]
    private void StartOfflineDemo()
    {
        _core.Dispatch("{\"kind\":\"new_offline_game\",\"payload\":{\"player_names\":[\"Voce\",\"CPU-2\"],\"cpu_flags\":[false,true],\"seed_lo\":7,\"seed_hi\":9}}");
        RefreshSnapshot();
        Status = "Rodando partida offline";
    }

    [RelayCommand]
    private void PlayCard(Card card)
    {
        if (card == null) return;
        _core.Dispatch($"{{\"kind\":\"play_card\",\"payload\":{{\"card\":{{\"Rank\":\"{card.Rank}\",\"Suit\":\"{card.Suit}\"}}}}}}");
    }

    [RelayCommand]
    private void RequestTruco()
    {
        _core.Dispatch("{\"kind\":\"request_truco\"}");
    }

    [RelayCommand]
    private void AcceptTruco()
    {
        _core.Dispatch("{\"kind\":\"accept_truco\"}");
    }

    [RelayCommand]
    private void RefuseTruco()
    {
        _core.Dispatch("{\"kind\":\"refuse_truco\"}");
    }
}
