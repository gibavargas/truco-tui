using System;
using System.IO;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Text.Json;

namespace TrucoWinUI.Services;

public sealed class TrucoCoreService : IDisposable
{
    private static readonly string DllName = "truco-core-ffi.dll";
    private static string? _cachedDllPath;
    private IntPtr _handle;

    static TrucoCoreService()
    {
        ExtractAndCacheDll();
    }

    private static void ExtractAndCacheDll()
    {
        if (_cachedDllPath != null && File.Exists(_cachedDllPath))
            return;

        var assembly = Assembly.GetExecutingAssembly();
        var resourceName = "TrucoWinUI.truco-core-ffi.dll";

        var tempPath = Path.Combine(Path.GetTempPath(), "TrucoWinUI");
        Directory.CreateDirectory(tempPath);
        _cachedDllPath = Path.Combine(tempPath, DllName);

        if (!File.Exists(_cachedDllPath))
        {
            using var stream = assembly.GetManifestResourceStream(resourceName);
            if (stream == null)
            {
                var binPath = Path.Combine(AppContext.BaseDirectory, DllName);
                if (File.Exists(binPath))
                {
                    _cachedDllPath = binPath;
                    return;
                }
                throw new InvalidOperationException($"Native DLL '{DllName}' not found. Please ensure the DLL is in the application directory.");
            }

            using var fileStream = File.Create(_cachedDllPath);
            stream.CopyTo(fileStream);
        }
    }

    private static class NativeMethods
    {
        [DllImport("kernel32", SetLastError = true, CharSet = CharSet.Unicode)]
        private static extern IntPtr LoadLibrary(string lpFileName);

        [DllImport("kernel32", SetLastError = true)]
        private static extern bool FreeLibrary(IntPtr hModule);

        [DllImport("kernel32", SetLastError = true, CharSet = CharSet.Ansi)]
        private static extern IntPtr GetProcAddress(IntPtr hModule, string lpProcName);

        private static IntPtr _libraryHandle;

        static NativeMethods()
        {
            _libraryHandle = LoadLibrary(_cachedDllPath!);
            if (_libraryHandle == IntPtr.Zero)
            {
                var error = Marshal.GetLastWin32Error();
                throw new DllNotFoundException($"Failed to load native library '{_cachedDllPath}'. Error code: {error}");
            }
        }

        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate IntPtr TrucoCoreCreateDelegate();
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate void TrucoCoreDestroyDelegate(IntPtr handle);
        [UnmanagedFunctionPointer(CallingConvention.Cdecl, CharSet = CharSet.Ansi)]
        public delegate IntPtr TrucoCoreDispatchIntentJSONDelegate(IntPtr handle, [MarshalAs(UnmanagedType.LPUTF8Str)] string payload);
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate IntPtr TrucoCorePollEventJSONDelegate(IntPtr handle);
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate IntPtr TrucoCoreSnapshotJSONDelegate(IntPtr handle);
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate void TrucoCoreFreeStringDelegate(IntPtr ptr);

        private static T GetDelegate<T>(IntPtr ptr) where T : Delegate
        {
            return (T)Marshal.GetDelegateForFunctionPointer(ptr, typeof(T));
        }

        public static TrucoCoreCreateDelegate TrucoCoreCreate => 
            GetDelegate<TrucoCoreCreateDelegate>(GetProcAddress(_libraryHandle, "TrucoCoreCreate"));
        public static TrucoCoreDestroyDelegate TrucoCoreDestroy => 
            GetDelegate<TrucoCoreDestroyDelegate>(GetProcAddress(_libraryHandle, "TrucoCoreDestroy"));
        public static TrucoCoreDispatchIntentJSONDelegate TrucoCoreDispatchIntentJSON => 
            GetDelegate<TrucoCoreDispatchIntentJSONDelegate>(GetProcAddress(_libraryHandle, "TrucoCoreDispatchIntentJSON"));
        public static TrucoCorePollEventJSONDelegate TrucoCorePollEventJSON => 
            GetDelegate<TrucoCorePollEventJSONDelegate>(GetProcAddress(_libraryHandle, "TrucoCorePollEventJSON"));
        public static TrucoCoreSnapshotJSONDelegate TrucoCoreSnapshotJSON => 
            GetDelegate<TrucoCoreSnapshotJSONDelegate>(GetProcAddress(_libraryHandle, "TrucoCoreSnapshotJSON"));
        public static TrucoCoreFreeStringDelegate TrucoCoreFreeString => 
            GetDelegate<TrucoCoreFreeStringDelegate>(GetProcAddress(_libraryHandle, "TrucoCoreFreeString"));
    }

    public TrucoCoreService()
    {
        _handle = NativeMethods.TrucoCoreCreate();
        if (_handle == IntPtr.Zero)
        {
            throw new InvalidOperationException("Failed to initialize TrucoCore - native library may be missing or incompatible");
        }
    }

    public string? Dispatch(string intentJson)
    {
        IntPtr resultPtr = NativeMethods.TrucoCoreDispatchIntentJSON(_handle, intentJson);
        return ReadAndFreeString(resultPtr);
    }

    public string? SnapshotJson()
    {
        IntPtr resultPtr = NativeMethods.TrucoCoreSnapshotJSON(_handle);
        return ReadAndFreeString(resultPtr);
    }

    public string? PollEventJson()
    {
        IntPtr resultPtr = NativeMethods.TrucoCorePollEventJSON(_handle);
        return ReadAndFreeString(resultPtr);
    }

    private string? ReadAndFreeString(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero) return null;
        string? result = Marshal.PtrToStringUTF8(ptr);
        NativeMethods.TrucoCoreFreeString(ptr);
        return result;
    }

    public void Dispose()
    {
        if (_handle != IntPtr.Zero)
        {
            NativeMethods.TrucoCoreDestroy(_handle);
            _handle = IntPtr.Zero;
        }
    }
}
