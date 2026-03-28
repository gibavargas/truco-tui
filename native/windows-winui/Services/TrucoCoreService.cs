using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text.Json;
using TrucoWinUI.Models;

namespace TrucoWinUI.Services;

public sealed class TrucoCoreService : IDisposable
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
    };

    private IntPtr _handle;

    public TrucoCoreService()
    {
        NativeDependencyValidator.EnsurePresent();
        _handle = NativeMethods.TrucoCoreCreate();
    }

    public SnapshotBundle GetSnapshot()
    {
        string json = ReadAndFreeString(NativeMethods.TrucoCoreSnapshotJSON(_handle))
            ?? throw new InvalidOperationException("Core returned an empty snapshot.");
        return JsonSerializer.Deserialize<SnapshotBundle>(json, JsonOptions)
            ?? new SnapshotBundle();
    }

    public AppEvent? PollEvent()
    {
        string? json = ReadAndFreeString(NativeMethods.TrucoCorePollEventJSON(_handle));
        if (string.IsNullOrWhiteSpace(json))
        {
            return null;
        }

        return JsonSerializer.Deserialize<AppEvent>(json, JsonOptions);
    }

    public CoreVersions GetVersions()
    {
        string json = ReadAndFreeString(NativeMethods.TrucoCoreVersionsJSON())
            ?? throw new InvalidOperationException("Core returned no version payload.");
        return JsonSerializer.Deserialize<CoreVersions>(json, JsonOptions)
            ?? new CoreVersions();
    }

    public AppError? SetLocale(string locale) => Dispatch("set_locale", new { locale });

    public AppError? NewHand() => Dispatch("new_hand", null);

    public AppError? StartOfflineGame(IReadOnlyList<string> playerNames, IReadOnlyList<bool> cpuFlags)
        => Dispatch("new_offline_game", new
        {
            player_names = playerNames,
            cpu_flags = cpuFlags,
        });

    public AppError? CreateHostSession(string hostName, int numPlayers, string? bindAddr, string? relayUrl, string? transportMode)
        => Dispatch("create_host_session", new
        {
            bind_addr = bindAddr ?? string.Empty,
            host_name = hostName,
            num_players = numPlayers,
            relay_url = relayUrl ?? string.Empty,
            transport_mode = transportMode ?? string.Empty,
        });

    public AppError? StartHostedMatch() => Dispatch("start_hosted_match", null);

    public AppError? JoinSession(string key, string playerName, string desiredRole)
        => Dispatch("join_session", new
        {
            key,
            player_name = playerName,
            desired_role = desiredRole,
        });

    public AppError? SendChat(string text) => Dispatch("send_chat", new { text });

    public AppError? VoteHost(int candidateSeat) => Dispatch("vote_host", new { candidate_seat = candidateSeat });

    public AppError? RequestReplacementInvite(int targetSeat)
        => Dispatch("request_replacement_invite", new { target_seat = targetSeat });

    public AppError? CloseSession() => Dispatch("close_session", null);

    public AppError? ResetSession() => Dispatch("reset", null);

    public AppError? PlayCard(int cardIndex, bool faceDown = false) => DispatchGameAction("play", cardIndex, faceDown);

    public AppError? RequestTruco() => DispatchGameAction("truco", 0);

    public AppError? AcceptTruco() => DispatchGameAction("accept", 0);

    public AppError? RefuseTruco() => DispatchGameAction("refuse", 0);

    public void Dispose()
    {
        if (_handle != IntPtr.Zero)
        {
            NativeMethods.TrucoCoreDestroy(_handle);
            _handle = IntPtr.Zero;
        }
    }

    private AppError? DispatchGameAction(string action, int cardIndex, bool faceDown = false)
        => Dispatch("game_action", new { action, card_index = cardIndex, face_down = faceDown });

    private AppError? Dispatch(string kind, object? payload)
    {
        if (_handle == IntPtr.Zero)
        {
            return new AppError { Code = "disposed", Message = "The native runtime has already been disposed." };
        }

        string intentJson = payload is null
            ? JsonSerializer.Serialize(new { kind }, JsonOptions)
            : JsonSerializer.Serialize(new { kind, payload }, JsonOptions);

        string? responseJson = ReadAndFreeString(NativeMethods.TrucoCoreDispatchIntentJSON(_handle, intentJson));
        if (string.IsNullOrWhiteSpace(responseJson))
        {
            return null;
        }

        return JsonSerializer.Deserialize<AppError>(responseJson, JsonOptions)
            ?? new AppError { Code = "dispatch_failed", Message = responseJson };
    }

    private static string? ReadAndFreeString(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero)
        {
            return null;
        }

        string? result = Marshal.PtrToStringUTF8(ptr);
        NativeMethods.TrucoCoreFreeString(ptr);
        return result;
    }

    private static class NativeMethods
    {
        private const string DllName = "truco-core-ffi";

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern IntPtr TrucoCoreCreate();

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern void TrucoCoreDestroy(IntPtr handle);

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl, CharSet = CharSet.Ansi)]
        public static extern IntPtr TrucoCoreDispatchIntentJSON(IntPtr handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string payload);

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern IntPtr TrucoCorePollEventJSON(IntPtr handle);

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern IntPtr TrucoCoreSnapshotJSON(IntPtr handle);

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern IntPtr TrucoCoreVersionsJSON();

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl)]
        public static extern void TrucoCoreFreeString(IntPtr ptr);
    }
}
