// 1. Get current PID and TID (upper 32 = PID, lower 32 = TID)
u64 id = bpf_get_current_pid_tgid();
u32 pid = id >> 32;

// 2. Get current UID (upper 32 = GID, lower 32 = UID)
u32 uid = (u32)bpf_get_current_uid_gid();

// 3. Get process name (up to 16 bytes, null-terminated)
char comm[16];
bpf_get_current_comm(&comm, sizeof(comm));

// 4. Read a string from userspace (e.g. a filename pointer)
char path[256];
bpf_probe_read_user_str(path, sizeof(path), user_ptr);

// 5. Reserve a slot in the ring buffer
struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
if (!e) return 0; // verifier requires this null check

// 6. Submit the slot
bpf_ringbuf_submit(e, 0);
