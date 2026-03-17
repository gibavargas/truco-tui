using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace TrucoWPF.Services;

public sealed class TrucoCoreService : IDisposable
{
    private const string DllName = "truco-core-ffi.dll";
    private const int RequiredCoreApiVersion = 1;
    private const int RequiredSnapshotSchemaVersion = 1;
    private static readonly Lazy<NativeBridge> SharedBridge = new(CreateBridge, isThreadSafe: true);

    private IntPtr _runtimeHandle;

    public TrucoCoreService()
    {
        _runtimeHandle = SharedBridge.Value.TrucoCoreCreate();
        if (_runtimeHandle == IntPtr.Zero)
        {
            throw new InvalidOperationException("Failed to initialize the Truco runtime.");
        }

        Versions = SharedBridge.Value.GetVersions();
        if (Versions.CoreApiVersion != RequiredCoreApiVersion ||
            Versions.SnapshotSchemaVersion != RequiredSnapshotSchemaVersion)
        {
            throw new InvalidOperationException(
                $"Incompatible Truco runtime at '{LibraryPath}'. " +
                $"Expected core_api={RequiredCoreApiVersion}, snapshot_schema={RequiredSnapshotSchemaVersion}; " +
                $"found core_api={Versions.CoreApiVersion}, snapshot_schema={Versions.SnapshotSchemaVersion}.");
        }
    }

    public CoreVersionsInfo Versions { get; }

    public string LibraryPath => SharedBridge.Value.LibraryPath;

    public string? Dispatch(string intentJson)
    {
        IntPtr resultPtr = SharedBridge.Value.TrucoCoreDispatchIntentJson(_runtimeHandle, intentJson);
        return ReadAndFreeString(resultPtr);
    }

    public string? SnapshotJson()
    {
        IntPtr resultPtr = SharedBridge.Value.TrucoCoreSnapshotJson(_runtimeHandle);
        return ReadAndFreeString(resultPtr);
    }

    public string? PollEventJson()
    {
        IntPtr resultPtr = SharedBridge.Value.TrucoCorePollEventJson(_runtimeHandle);
        return ReadAndFreeString(resultPtr);
    }

    private static NativeBridge CreateBridge()
    {
        List<string> candidates = EnumerateCandidatePaths()
            .Distinct(StringComparer.OrdinalIgnoreCase)
            .ToList();
        List<string> existing = candidates
            .Where(File.Exists)
            .ToList();

        if (existing.Count == 0)
        {
            throw new DllNotFoundException(
                $"Unable to find '{DllName}'. Checked:{Environment.NewLine}{string.Join(Environment.NewLine, candidates.Select(path => $" - {path}"))}");
        }

        List<string> loadErrors = new();
        foreach (string candidate in existing)
        {
            IntPtr libraryHandle = LoadLibraryEx(candidate, IntPtr.Zero, LoadLibrarySearchDllLoadDir | LoadLibrarySearchDefaultDirs);
            if (libraryHandle == IntPtr.Zero)
            {
                int error = Marshal.GetLastWin32Error();
                loadErrors.Add($"{candidate} (Win32 {error})");
                continue;
            }

            try
            {
                return new NativeBridge(
                    libraryHandle,
                    candidate,
                    ResolveDelegate<TrucoCoreCreateDelegate>(libraryHandle, "TrucoCoreCreate"),
                    ResolveDelegate<TrucoCoreDestroyDelegate>(libraryHandle, "TrucoCoreDestroy"),
                    ResolveDelegate<TrucoCoreDispatchIntentJsonDelegate>(libraryHandle, "TrucoCoreDispatchIntentJSON"),
                    ResolveDelegate<TrucoCorePollEventJsonDelegate>(libraryHandle, "TrucoCorePollEventJSON"),
                    ResolveDelegate<TrucoCoreSnapshotJsonDelegate>(libraryHandle, "TrucoCoreSnapshotJSON"),
                    ResolveDelegate<TrucoCoreVersionsJsonDelegate>(libraryHandle, "TrucoCoreVersionsJSON"),
                    ResolveDelegate<TrucoCoreFreeStringDelegate>(libraryHandle, "TrucoCoreFreeString"));
            }
            catch
            {
                FreeLibrary(libraryHandle);
                throw;
            }
        }

        throw new DllNotFoundException(
            $"Found '{DllName}' but failed to load it from any app-local path:{Environment.NewLine}{string.Join(Environment.NewLine, loadErrors.Select(error => $" - {error}"))}");
    }

    private static IEnumerable<string> EnumerateCandidatePaths()
    {
        if (!string.IsNullOrWhiteSpace(Environment.GetEnvironmentVariable("TRUCO_CORE_LIB")))
        {
            yield return Path.GetFullPath(Environment.GetEnvironmentVariable("TRUCO_CORE_LIB")!);
        }

        string baseDirectory = AppContext.BaseDirectory;
        yield return Path.Combine(baseDirectory, DllName);
        yield return Path.Combine(baseDirectory, "lib", DllName);
        yield return Path.Combine(baseDirectory, "runtimes", "win-x64", "native", DllName);

        string currentDirectory = Environment.CurrentDirectory;
        yield return Path.Combine(currentDirectory, DllName);
        yield return Path.Combine(currentDirectory, "lib", DllName);
        yield return Path.Combine(currentDirectory, "bin", DllName);
        yield return Path.Combine(currentDirectory, "native", "windows-wpf", "lib", DllName);
    }

    private static T ResolveDelegate<T>(IntPtr libraryHandle, string exportName) where T : Delegate
    {
        IntPtr functionPtr = GetProcAddress(libraryHandle, exportName);
        if (functionPtr == IntPtr.Zero)
        {
            throw new MissingMethodException($"Export '{exportName}' was not found in '{DllName}'.");
        }

        return Marshal.GetDelegateForFunctionPointer<T>(functionPtr);
    }

    private string? ReadAndFreeString(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero)
        {
            return null;
        }

        string? result = Marshal.PtrToStringUTF8(ptr);
        SharedBridge.Value.TrucoCoreFreeString(ptr);
        return result;
    }

    public void Dispose()
    {
        if (_runtimeHandle == IntPtr.Zero)
        {
            return;
        }

        SharedBridge.Value.TrucoCoreDestroy(_runtimeHandle);
        _runtimeHandle = IntPtr.Zero;
    }

    private sealed record NativeBridge(
        IntPtr LibraryHandle,
        string LibraryPath,
        TrucoCoreCreateDelegate TrucoCoreCreate,
        TrucoCoreDestroyDelegate TrucoCoreDestroy,
        TrucoCoreDispatchIntentJsonDelegate TrucoCoreDispatchIntentJson,
        TrucoCorePollEventJsonDelegate TrucoCorePollEventJson,
        TrucoCoreSnapshotJsonDelegate TrucoCoreSnapshotJson,
        TrucoCoreVersionsJsonDelegate TrucoCoreVersionsJson,
        TrucoCoreFreeStringDelegate TrucoCoreFreeString)
    {
        public CoreVersionsInfo GetVersions()
        {
            IntPtr resultPtr = TrucoCoreVersionsJson();
            if (resultPtr == IntPtr.Zero)
            {
                throw new InvalidOperationException("The runtime did not return version metadata.");
            }

            try
            {
                string? json = Marshal.PtrToStringUTF8(resultPtr);
                CoreVersionsInfo? versions = JsonSerializer.Deserialize<CoreVersionsInfo>(json ?? string.Empty);
                if (versions == null)
                {
                    throw new InvalidOperationException("The runtime returned unreadable version metadata.");
                }

                return versions;
            }
            finally
            {
                TrucoCoreFreeString(resultPtr);
            }
        }
    }

    public sealed class CoreVersionsInfo
    {
        [JsonPropertyName("core_api_version")]
        public int CoreApiVersion { get; init; }

        [JsonPropertyName("protocol_version")]
        public int ProtocolVersion { get; init; }

        [JsonPropertyName("snapshot_schema_version")]
        public int SnapshotSchemaVersion { get; init; }
    }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreCreateDelegate();

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void TrucoCoreDestroyDelegate(IntPtr handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl, CharSet = CharSet.Ansi)]
    private delegate IntPtr TrucoCoreDispatchIntentJsonDelegate(
        IntPtr handle,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string payload);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCorePollEventJsonDelegate(IntPtr handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreSnapshotJsonDelegate(IntPtr handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreVersionsJsonDelegate();

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void TrucoCoreFreeStringDelegate(IntPtr ptr);

    private const uint LoadLibrarySearchDllLoadDir = 0x00000100;
    private const uint LoadLibrarySearchDefaultDirs = 0x00001000;

    [DllImport("kernel32", SetLastError = true, CharSet = CharSet.Unicode)]
    private static extern IntPtr LoadLibraryEx(string lpFileName, IntPtr hFile, uint dwFlags);

    [DllImport("kernel32", SetLastError = true)]
    private static extern bool FreeLibrary(IntPtr hModule);

    [DllImport("kernel32", SetLastError = true, CharSet = CharSet.Ansi)]
    private static extern IntPtr GetProcAddress(IntPtr hModule, string lpProcName);
}
