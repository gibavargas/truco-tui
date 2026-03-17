using System.Reflection;
using System.Security.Cryptography;

namespace TrucoWinUI.Services;

internal static class TrucoCoreLibraryLocator
{
    private const string DllName = "truco-core-ffi.dll";
    private const string ResourceName = "TrucoWinUI.truco-core-ffi.dll";

    public static string ResolveLibraryPath()
    {
        var candidates = CandidateLibraryPaths().ToList();
        var existing = candidates.FirstOrDefault(File.Exists);
        if (!string.IsNullOrWhiteSpace(existing))
        {
            return existing;
        }

        var extracted = ExtractEmbeddedLibrary();
        if (!string.IsNullOrWhiteSpace(extracted))
        {
            return extracted;
        }

        throw new InvalidOperationException(
            $"Unable to locate {DllName}. Checked: {string.Join(", ", candidates)}");
    }

    public static IEnumerable<string> CandidateLibraryPaths()
    {
        var paths = new List<string>();

        var explicitOverride = Environment.GetEnvironmentVariable("TRUCO_CORE_LIB");
        if (!string.IsNullOrWhiteSpace(explicitOverride))
        {
            paths.Add(Path.GetFullPath(explicitOverride));
        }

        var baseDirectory = AppContext.BaseDirectory;
        paths.Add(Path.Combine(baseDirectory, DllName));
        paths.Add(Path.Combine(baseDirectory, "lib", DllName));

        var currentDirectory = Environment.CurrentDirectory;
        paths.Add(Path.Combine(currentDirectory, DllName));
        paths.Add(Path.Combine(currentDirectory, "bin", DllName));
        paths.Add(Path.Combine(currentDirectory, "native", "windows-winui", DllName));

        return paths
            .Where(path => !string.IsNullOrWhiteSpace(path))
            .Distinct(StringComparer.OrdinalIgnoreCase);
    }

    private static string? ExtractEmbeddedLibrary()
    {
        var assembly = Assembly.GetExecutingAssembly();
        using var stream = assembly.GetManifestResourceStream(ResourceName);
        if (stream == null)
        {
            return null;
        }

        using var ms = new MemoryStream();
        stream.CopyTo(ms);
        var payload = ms.ToArray();
        var hash = Convert.ToHexString(SHA256.HashData(payload)).ToLowerInvariant().Substring(0, 12);
        var version = assembly.GetName().Version?.ToString() ?? "dev";
        var cacheDirectory = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
            "Truco",
            "runtime-cache",
            "winui",
            version,
            hash);

        Directory.CreateDirectory(cacheDirectory);
        var outputPath = Path.Combine(cacheDirectory, DllName);
        if (!File.Exists(outputPath))
        {
            File.WriteAllBytes(outputPath, payload);
        }

        return outputPath;
    }
}
