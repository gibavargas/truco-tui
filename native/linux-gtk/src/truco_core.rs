use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::sync::Arc;

struct CoreInstance {
    handle: usize,
    lib: libloading::Library,
}

impl Drop for CoreInstance {
    fn drop(&mut self) {
        unsafe {
            if let Ok(func) = self.lib.get::<unsafe extern "C" fn(usize)>(b"TrucoCoreDestroy") {
                func(self.handle);
            }
        }
    }
}

#[derive(Clone)]
pub struct TrucoCore {
    inner: Arc<CoreInstance>,
}

impl TrucoCore {
    pub fn new() -> Self {
        unsafe {
            let lib = libloading::Library::new("libtruco-core-ffi.so").expect("Failed to load libtruco-core-ffi.so - Ensure it's compiled and in PATH/LD_LIBRARY_PATH");
            let create: libloading::Symbol<unsafe extern "C" fn() -> usize> = lib.get(b"TrucoCoreCreate").unwrap();
            let handle = create();
            
            Self {
                inner: Arc::new(CoreInstance { handle, lib }),
            }
        }
    }

    pub fn dispatch(&self, intent_json: &str) -> Option<String> {
        let c_intent = CString::new(intent_json).unwrap();
        unsafe {
            let func: libloading::Symbol<unsafe extern "C" fn(usize, *const c_char) -> *mut c_char> = self.inner.lib.get(b"TrucoCoreDispatchIntentJSON").unwrap();
            let res_ptr = func(self.inner.handle, c_intent.as_ptr());
            self.read_and_free(res_ptr)
        }
    }

    pub fn poll_event(&self) -> Option<String> {
        unsafe {
            let func: libloading::Symbol<unsafe extern "C" fn(usize) -> *mut c_char> = self.inner.lib.get(b"TrucoCorePollEventJSON").unwrap();
            let res_ptr = func(self.inner.handle);
            self.read_and_free(res_ptr)
        }
    }

    pub fn snapshot(&self) -> Option<String> {
        unsafe {
            let func: libloading::Symbol<unsafe extern "C" fn(usize) -> *mut c_char> = self.inner.lib.get(b"TrucoCoreSnapshotJSON").unwrap();
            let res_ptr = func(self.inner.handle);
            self.read_and_free(res_ptr)
        }
    }

    fn read_and_free(&self, ptr: *mut c_char) -> Option<String> {
        if ptr.is_null() {
            return None;
        }
        unsafe {
            let c_str = CStr::from_ptr(ptr);
            let result = c_str.to_string_lossy().into_owned();
            let free_str: libloading::Symbol<unsafe extern "C" fn(*mut c_char)> = self.inner.lib.get(b"TrucoCoreFreeString").unwrap();
            free_str(ptr);
            Some(result)
        }
    }
}
