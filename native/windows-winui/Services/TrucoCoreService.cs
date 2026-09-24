using System;
using System.Collections.Generic;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Text.Json;
using TrucoWinUI.Contracts;
using TrucoWinUI.Models;

namespace TrucoWinUI.Services;

public sealed class TrucoCoreService : IDisposable
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
    };

    private IntPtr _handle;

    static TrucoCoreService()
    {
        NativeLibrary.SetDllImportResolver(typeof(TrucoCoreService).Assembly, ResolveNativeLibrary);
    }

    public TrucoCoreService()
    {
        string libraryPath = TrucoCoreLibraryLocator.ResolveLibraryPath();
        NativeDependencyValidator.EnsurePresent(libraryPath);
        EnsureCompatibleVersions(GetVersions());
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

    public AppError? SetLocale(string locale) => Dispatch(RuntimeContract.SetLocale, new { locale });

    public AppError? NewHand() => Dispatch(RuntimeContract.NewHand, null);

    public AppError? Tick(int maxSteps = 12) => Dispatch(RuntimeContract.Tick, new { max_steps = maxSteps });

    public AppError? StartOfflineGame(IReadOnlyList<string> playerNames, IReadOnlyList<bool> cpuFlags)
        => Dispatch(RuntimeContract.NewOfflineGame, new
        {
            player_names = playerNames,
            cpu_flags = cpuFlags,
        });

    public AppError? CreateHostSession(string hostName, int numPlayers, string? bindAddr, string? relayUrl, string? transportMode)
        => Dispatch(RuntimeContract.CreateHostSession, new
        {
            bind_addr = bindAddr ?? string.Empty,
            host_name = hostName,
            num_players = numPlayers,
            relay_url = relayUrl ?? string.Empty,
            transport_mode = transportMode ?? string.Empty,
        });

    public AppError? StartHostedMatch() => Dispatch(RuntimeContract.StartHostedMatch, null);

    public AppError? JoinSession(string key, string playerName, string desiredRole)
        => Dispatch(RuntimeContract.JoinSession, new
        {
            key,
            player_name = playerName,
            desired_role = desiredRole,
        });

    public AppError? SendChat(string text) => Dispatch(RuntimeContract.SendChat, new { text });

    public AppError? VoteHost(int candidateSeat) => Dispatch(RuntimeContract.VoteHost, new { candidate_seat = candidateSeat });

    public AppError? RequestReplacementInvite(int targetSeat)
        => Dispatch(RuntimeContract.RequestReplacementInvite, new { target_seat = targetSeat });

    public AppError? CloseSession() => Dispatch(RuntimeContract.CloseSession, null);

    public AppError? ResetSession() => CloseSession();

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
        => Dispatch(RuntimeContract.GameAction, new { action, card_index = cardIndex, face_down = faceDown });

    private static IntPtr ResolveNativeLibrary(string libraryName, Assembly assembly, DllImportSearchPath? searchPath)
    {
        if (!IsCoreLibraryName(libraryName))
        {
            return IntPtr.Zero;
        }

        string libraryPath = TrucoCoreLibraryLocator.ResolveLibraryPath();
        NativeDependencyValidator.EnsurePresent(libraryPath);
        return NativeLibrary.Load(libraryPath);
    }

    private static bool IsCoreLibraryName(string libraryName)
        => string.Equals(libraryName, "truco-core-ffi", StringComparison.OrdinalIgnoreCase)
           || string.Equals(libraryName, "truco-core-ffi.dll", StringComparison.OrdinalIgnoreCase);

    private static void EnsureCompatibleVersions(CoreVersions versions)
    {
        if (versions.CoreApiVersion != RuntimeContract.CoreApiVersion ||
            versions.ProtocolVersion != RuntimeContract.ProtocolVersion ||
            versions.SnapshotSchemaVersion != RuntimeContract.SnapshotSchemaVersion)
        {
            throw new InvalidOperationException(
                "Incompatible Truco core runtime. " +
                $"Expected core API {RuntimeContract.CoreApiVersion}, protocol {RuntimeContract.ProtocolVersion}, snapshot {RuntimeContract.SnapshotSchemaVersion}; " +
                $"got core API {versions.CoreApiVersion}, protocol {versions.ProtocolVersion}, snapshot {versions.SnapshotSchemaVersion}.");
        }
    }

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

        [DllImport(DllName, CallingConvention = CallingConvention.Cdecl, CharSet = CharSet.Ansi)]
        public static extern IntPtr TrucoCoreCreateWithConfigJSON([MarshalAs(UnmanagedType.LPUTF8Str)] string payload);

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
