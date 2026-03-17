using System.Runtime.InteropServices;
using System.Text.Json;
using TrucoWinUI.Models;

namespace TrucoWinUI.Services;

public sealed class TrucoCoreService : IDisposable
{
    private const int RequiredCoreApiVersion = 1;
    private const int RequiredSnapshotSchemaVersion = 1;

    private static readonly Lazy<NativeBindings> Bindings = new(NativeBindings.Load, true);

    private nuint _handle;

    public TrucoCoreService()
    {
        _handle = Bindings.Value.TrucoCoreCreate();
        if (_handle == 0)
        {
            throw new InvalidOperationException("Failed to initialize the shared Truco runtime.");
        }
    }

    public string LibraryPath => Bindings.Value.LibraryPath;

    public CoreVersions Versions => Bindings.Value.Versions;

    public string? Dispatch(string intentJson)
    {
        return ReadAndFreeString(Bindings.Value.TrucoCoreDispatchIntentJson(_handle, intentJson));
    }

    public string? SnapshotJson()
    {
        return ReadAndFreeString(Bindings.Value.TrucoCoreSnapshotJson(_handle));
    }

    public string? PollEventJson()
    {
        return ReadAndFreeString(Bindings.Value.TrucoCorePollEventJson(_handle));
    }

    public void Dispose()
    {
        if (_handle == 0)
        {
            return;
        }

        Bindings.Value.TrucoCoreDestroy(_handle);
        _handle = 0;
    }

    private static string? ReadAndFreeString(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero)
        {
            return null;
        }

        try
        {
            return Marshal.PtrToStringUTF8(ptr);
        }
        finally
        {
            Bindings.Value.TrucoCoreFreeString(ptr);
        }
    }

    private sealed class NativeBindings
    {
        private readonly nint _libraryHandle;

        private NativeBindings(
            nint libraryHandle,
            string libraryPath,
            CoreVersions versions,
            TrucoCoreCreateDelegate trucoCoreCreate,
            TrucoCoreDestroyDelegate trucoCoreDestroy,
            TrucoCoreDispatchIntentJsonDelegate trucoCoreDispatchIntentJson,
            TrucoCorePollEventJsonDelegate trucoCorePollEventJson,
            TrucoCoreSnapshotJsonDelegate trucoCoreSnapshotJson,
            TrucoCoreVersionsJsonDelegate trucoCoreVersionsJson,
            TrucoCoreFreeStringDelegate trucoCoreFreeString)
        {
            _libraryHandle = libraryHandle;
            LibraryPath = libraryPath;
            Versions = versions;
            TrucoCoreCreate = trucoCoreCreate;
            TrucoCoreDestroy = trucoCoreDestroy;
            TrucoCoreDispatchIntentJson = trucoCoreDispatchIntentJson;
            TrucoCorePollEventJson = trucoCorePollEventJson;
            TrucoCoreSnapshotJson = trucoCoreSnapshotJson;
            TrucoCoreVersionsJson = trucoCoreVersionsJson;
            TrucoCoreFreeString = trucoCoreFreeString;
        }

        public string LibraryPath { get; }

        public CoreVersions Versions { get; }

        public TrucoCoreCreateDelegate TrucoCoreCreate { get; }

        public TrucoCoreDestroyDelegate TrucoCoreDestroy { get; }

        public TrucoCoreDispatchIntentJsonDelegate TrucoCoreDispatchIntentJson { get; }

        public TrucoCorePollEventJsonDelegate TrucoCorePollEventJson { get; }

        public TrucoCoreSnapshotJsonDelegate TrucoCoreSnapshotJson { get; }

        public TrucoCoreVersionsJsonDelegate TrucoCoreVersionsJson { get; }

        public TrucoCoreFreeStringDelegate TrucoCoreFreeString { get; }

        public static NativeBindings Load()
        {
            var libraryPath = TrucoCoreLibraryLocator.ResolveLibraryPath();
            var libraryHandle = NativeLibrary.Load(libraryPath);

            var bindings = new NativeBindings(
                libraryHandle,
                libraryPath,
                new CoreVersions(),
                GetExport<TrucoCoreCreateDelegate>(libraryHandle, "TrucoCoreCreate"),
                GetExport<TrucoCoreDestroyDelegate>(libraryHandle, "TrucoCoreDestroy"),
                GetExport<TrucoCoreDispatchIntentJsonDelegate>(libraryHandle, "TrucoCoreDispatchIntentJSON"),
                GetExport<TrucoCorePollEventJsonDelegate>(libraryHandle, "TrucoCorePollEventJSON"),
                GetExport<TrucoCoreSnapshotJsonDelegate>(libraryHandle, "TrucoCoreSnapshotJSON"),
                GetExport<TrucoCoreVersionsJsonDelegate>(libraryHandle, "TrucoCoreVersionsJSON"),
                GetExport<TrucoCoreFreeStringDelegate>(libraryHandle, "TrucoCoreFreeString"));

            var versions = bindings.ReadVersions();
            if (versions.CoreApiVersion != RequiredCoreApiVersion ||
                versions.SnapshotSchemaVersion != RequiredSnapshotSchemaVersion)
            {
                throw new InvalidOperationException(
                    $"Incompatible Truco runtime at '{libraryPath}'. " +
                    $"Expected core_api={RequiredCoreApiVersion}, snapshot_schema={RequiredSnapshotSchemaVersion}; " +
                    $"found core_api={versions.CoreApiVersion}, snapshot_schema={versions.SnapshotSchemaVersion}.");
            }

            return new NativeBindings(
                libraryHandle,
                libraryPath,
                versions,
                bindings.TrucoCoreCreate,
                bindings.TrucoCoreDestroy,
                bindings.TrucoCoreDispatchIntentJson,
                bindings.TrucoCorePollEventJson,
                bindings.TrucoCoreSnapshotJson,
                bindings.TrucoCoreVersionsJson,
                bindings.TrucoCoreFreeString);
        }

        private CoreVersions ReadVersions()
        {
            var ptr = TrucoCoreVersionsJson();
            if (ptr == IntPtr.Zero)
            {
                throw new InvalidOperationException(
                    $"Shared runtime '{LibraryPath}' did not expose version metadata.");
            }

            var json = ReadAndFreeOwnedString(ptr);
            if (string.IsNullOrWhiteSpace(json))
            {
                throw new InvalidOperationException(
                    $"Shared runtime '{LibraryPath}' returned an empty version payload.");
            }

            var versions = JsonSerializer.Deserialize<CoreVersions>(json, JsonOptions.Default);
            if (versions == null)
            {
                throw new InvalidOperationException(
                    $"Shared runtime '{LibraryPath}' returned malformed version JSON.");
            }

            return versions;
        }

        private string? ReadAndFreeOwnedString(IntPtr ptr)
        {
            if (ptr == IntPtr.Zero)
            {
                return null;
            }

            try
            {
                return Marshal.PtrToStringUTF8(ptr);
            }
            finally
            {
                TrucoCoreFreeString(ptr);
            }
        }

        private static T GetExport<T>(nint libraryHandle, string symbolName) where T : Delegate
        {
            if (!NativeLibrary.TryGetExport(libraryHandle, symbolName, out var symbol))
            {
                throw new MissingMethodException($"Shared runtime is missing required symbol '{symbolName}'.");
            }

            return Marshal.GetDelegateForFunctionPointer<T>(symbol);
        }
    }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate nuint TrucoCoreCreateDelegate();

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void TrucoCoreDestroyDelegate(nuint handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreSnapshotJsonDelegate(nuint handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCorePollEventJsonDelegate(nuint handle);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreVersionsJsonDelegate();

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void TrucoCoreFreeStringDelegate(IntPtr ptr);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr TrucoCoreDispatchIntentJsonDelegate(
        nuint handle,
        [MarshalAs(UnmanagedType.LPUTF8Str)] string payload);
}
