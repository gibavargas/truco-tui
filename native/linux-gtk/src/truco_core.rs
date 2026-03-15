use crate::models::CoreVersions;
use libloading::{Library, Symbol};
use std::env;
use std::ffi::{CStr, CString};
use std::fmt;
use std::os::raw::c_char;
use std::path::{Path, PathBuf};
use std::sync::Arc;

const REQUIRED_CORE_API_VERSION: i32 = 1;
const REQUIRED_SNAPSHOT_SCHEMA_VERSION: i32 = 1;

struct CoreInstance {
    handle: usize,
    lib: Library,
    library_path: PathBuf,
}

impl Drop for CoreInstance {
    fn drop(&mut self) {
        unsafe {
            if let Ok(func) = self
                .lib
                .get::<unsafe extern "C" fn(usize)>(b"TrucoCoreDestroy")
            {
                func(self.handle);
            }
        }
    }
}

#[derive(Clone)]
pub struct TrucoCore {
    inner: Arc<CoreInstance>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CoreError {
    LibraryNotFound(Vec<PathBuf>),
    LibraryLoad(String),
    VersionUnavailable,
    IncompatibleVersion(CoreVersions),
    InvalidUtf8,
    InvalidJson(String),
    SymbolMissing(&'static str),
}

impl fmt::Display for CoreError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::LibraryNotFound(paths) => write!(
                f,
                "libtruco_core.so not found. Checked: {}",
                paths
                    .iter()
                    .map(|p| p.display().to_string())
                    .collect::<Vec<_>>()
                    .join(", ")
            ),
            Self::LibraryLoad(err) => write!(f, "failed to load Linux runtime: {err}"),
            Self::VersionUnavailable => write!(f, "runtime did not expose version metadata"),
            Self::IncompatibleVersion(v) => write!(
                f,
                "runtime version mismatch: core_api={} snapshot_schema={}",
                v.core_api_version, v.snapshot_schema_version
            ),
            Self::InvalidUtf8 => write!(f, "runtime returned invalid UTF-8"),
            Self::InvalidJson(err) => write!(f, "runtime returned invalid JSON: {err}"),
            Self::SymbolMissing(symbol) => write!(f, "runtime symbol missing: {symbol}"),
        }
    }
}

impl TrucoCore {
    pub fn new() -> Result<Self, CoreError> {
        unsafe {
            let candidate_paths = candidate_library_paths();
            let library_path = candidate_paths
                .iter()
                .find(|path| path.exists())
                .cloned()
                .ok_or_else(|| CoreError::LibraryNotFound(candidate_paths.clone()))?;

            let lib = Library::new(&library_path)
                .map_err(|err| CoreError::LibraryLoad(err.to_string()))?;
            let create: Symbol<unsafe extern "C" fn() -> usize> = lib
                .get(b"TrucoCoreCreate")
                .map_err(|_| CoreError::SymbolMissing("TrucoCoreCreate"))?;
            let handle = create();
            let core = Self {
                inner: Arc::new(CoreInstance {
                    handle,
                    lib,
                    library_path,
                }),
            };
            let versions = core.versions()?;
            if versions.core_api_version != REQUIRED_CORE_API_VERSION
                || versions.snapshot_schema_version != REQUIRED_SNAPSHOT_SCHEMA_VERSION
            {
                return Err(CoreError::IncompatibleVersion(versions));
            }
            Ok(core)
        }
    }

    pub fn library_path(&self) -> &Path {
        &self.inner.library_path
    }

    pub fn dispatch(&self, intent_json: &str) -> Result<Option<String>, CoreError> {
        let c_intent =
            CString::new(intent_json).map_err(|err| CoreError::InvalidJson(err.to_string()))?;
        unsafe {
            let func: Symbol<unsafe extern "C" fn(usize, *const c_char) -> *mut c_char> = self
                .inner
                .lib
                .get(b"TrucoCoreDispatchIntentJSON")
                .map_err(|_| CoreError::SymbolMissing("TrucoCoreDispatchIntentJSON"))?;
            let res_ptr = func(self.inner.handle, c_intent.as_ptr());
            self.read_and_free(res_ptr)
        }
    }

    pub fn poll_event(&self) -> Result<Option<String>, CoreError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(usize) -> *mut c_char> = self
                .inner
                .lib
                .get(b"TrucoCorePollEventJSON")
                .map_err(|_| CoreError::SymbolMissing("TrucoCorePollEventJSON"))?;
            let res_ptr = func(self.inner.handle);
            self.read_and_free(res_ptr)
        }
    }

    pub fn snapshot(&self) -> Result<Option<String>, CoreError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn(usize) -> *mut c_char> = self
                .inner
                .lib
                .get(b"TrucoCoreSnapshotJSON")
                .map_err(|_| CoreError::SymbolMissing("TrucoCoreSnapshotJSON"))?;
            let res_ptr = func(self.inner.handle);
            self.read_and_free(res_ptr)
        }
    }

    pub fn versions(&self) -> Result<CoreVersions, CoreError> {
        unsafe {
            let func: Symbol<unsafe extern "C" fn() -> *mut c_char> = self
                .inner
                .lib
                .get(b"TrucoCoreVersionsJSON")
                .map_err(|_| CoreError::SymbolMissing("TrucoCoreVersionsJSON"))?;
            let res_ptr = func();
            let json = self
                .read_and_free(res_ptr)?
                .ok_or(CoreError::VersionUnavailable)?;
            serde_json::from_str(&json).map_err(|err| CoreError::InvalidJson(err.to_string()))
        }
    }

    fn read_and_free(&self, ptr: *mut c_char) -> Result<Option<String>, CoreError> {
        if ptr.is_null() {
            return Ok(None);
        }
        unsafe {
            let c_str = CStr::from_ptr(ptr);
            let result = c_str
                .to_str()
                .map_err(|_| CoreError::InvalidUtf8)?
                .to_owned();
            let free_str: Symbol<unsafe extern "C" fn(*mut c_char)> = self
                .inner
                .lib
                .get(b"TrucoCoreFreeString")
                .map_err(|_| CoreError::SymbolMissing("TrucoCoreFreeString"))?;
            free_str(ptr);
            Ok(Some(result))
        }
    }
}

pub fn candidate_library_paths() -> Vec<PathBuf> {
    let mut paths = Vec::new();
    if let Ok(explicit) = env::var("TRUCO_CORE_LIB") {
        paths.push(PathBuf::from(explicit));
    }

    if let Ok(cwd) = env::current_dir() {
        paths.push(cwd.join("bin/libtruco_core.so"));
        paths.push(cwd.join("native/linux-gtk/lib/libtruco_core.so"));
        paths.push(cwd.join("lib/libtruco_core.so"));
        paths.push(cwd.join("libtruco_core.so"));
    }

    if let Ok(exe) = env::current_exe() {
        if let Some(dir) = exe.parent() {
            paths.push(dir.join("libtruco_core.so"));
            paths.push(dir.join("../lib/libtruco_core.so"));
        }
    }

    dedupe_paths(paths)
}

fn dedupe_paths(paths: Vec<PathBuf>) -> Vec<PathBuf> {
    let mut out = Vec::new();
    for path in paths {
        if !out.iter().any(|existing| existing == &path) {
            out.push(path);
        }
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::{SystemTime, UNIX_EPOCH};

    #[test]
    fn candidate_paths_include_explicit_override() {
        let suffix = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("time")
            .as_nanos();
        let candidate = env::temp_dir().join(format!("truco-linux-test-{suffix}.so"));
        env::set_var("TRUCO_CORE_LIB", &candidate);
        let paths = candidate_library_paths();
        env::remove_var("TRUCO_CORE_LIB");
        assert_eq!(paths.first(), Some(&candidate));
    }
}
