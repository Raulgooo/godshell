// ebpf/events.h
struct event {
**u64 ts; // bpf_ktime_get_ns()
**u32 pid;
**u32 uid;
char comm[16];
char path[256]; // only for openat events
**u8 type; // 0=exec 1=open 2=exit 3=connect
};
