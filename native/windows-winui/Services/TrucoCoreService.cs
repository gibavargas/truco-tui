using System;
using System.Runtime.InteropServices;
using System.Text.Json;

namespace TrucoWinUI.Services;

public sealed class TrucoCoreService : IDisposable
{
    private IntPtr _handle;

    public TrucoCoreService()
    {
        _handle = NativeMethods.TrucoCoreCreate();
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
        public static extern void TrucoCoreFreeString(IntPtr ptr);
    }
}
