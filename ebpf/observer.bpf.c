// go:build ignore

#include "common.h"

char __license[] SEC("license") = "Dual MIT/GPL";

/* ── Event types ─────────────────────────────────────────────────────────── */

#define EVENT_EXEC 0
#define EVENT_OPEN 1
#define EVENT_EXIT 2
#define EVENT_CONNECT 3

/* ── Shared event struct (kernel → userspace via ring buffer) ────────────── *
 *
 * Fixed-size. Go reads this byte-for-byte off the ring buffer and deserializes
 * it into the matching observer.Event struct (layout must be identical).
 *
 * path is reused for:
 *   EVENT_OPEN    → the file path passed to openat()
 *   EVENT_EXEC    → the executable filename
 *   EVENT_EXIT    → unused
 *   EVENT_CONNECT → unused (comm alone identifies the process)
 */
struct event {
  __u64 ts; /* bpf_ktime_get_ns() — nanoseconds since boot */
  __u32 pid;
  __u32 uid;
  __u8 type;      /* EVENT_* constant above                       */
  char comm[16];  /* process name, up to 16 bytes                 */
  char path[256]; /* file path or executable name                 */
};

/* ── Ring buffer map ─────────────────────────────────────────────────────── *
 *
 * Single ring buffer shared by all four handlers. Go reads from it via
 * ringbuf.NewReader(objs.Events).
 * max_entries = buffer size in bytes; must be a power-of-2 multiple of 4096.
 */
struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 256 * 1024); /* 256 KB */
} events SEC(".maps");

/* ── Tracepoint context structs ──────────────────────────────────────────── *
 *
 * These mirror the format files under /sys/kernel/tracing/events/.
 * Only the fields we actually use are declared; padding fills the rest.
 */

/* /sys/kernel/tracing/events/sched/sched_process_exec/format */
struct sched_exec_ctx {
  unsigned short common_type;
  unsigned char common_flags;
  unsigned char common_preempt_count;
  int common_pid;
  int __data_loc_filename; /* __data_loc — do not dereference directly */
  int pid;
  int old_pid;
};

/* /sys/kernel/tracing/events/sched/sched_process_exit/format */
struct sched_exit_ctx {
  unsigned short common_type;
  unsigned char common_flags;
  unsigned char common_preempt_count;
  int common_pid;
  char comm[16];
  int pid;
  int prio;
};

/* /sys/kernel/tracing/events/syscalls/sys_enter_openat/format
 * /sys/kernel/tracing/events/syscalls/sys_enter_connect/format
 * Both are sys_enter tracepoints: args[0..5] = syscall arguments. */
struct sys_enter_ctx {
  unsigned short common_type;
  unsigned char common_flags;
  unsigned char common_preempt_count;
  int common_pid;
  long __syscall_nr;
  unsigned long args[6];
};

/* ── Helper ──────────────────────────────────────────────────────────────── */

static __always_inline __u64 now(void) { return bpf_ktime_get_ns(); }

/* ── Handlers ────────────────────────────────────────────────────────────── */

/*
 * sched_process_exec — fires after execve() succeeds.
 * We get the real comm of the new process via bpf_get_current_comm().
 * The __data_loc filename field requires vmlinux.h to decode cleanly;
 * for v0.1 comm is enough (it IS the binary name after exec succeeds).
 */
SEC("tracepoint/sched/sched_process_exec")
int trace_exec(struct sched_exec_ctx *ctx) {
  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_EXEC;
  e->pid = (__u32)ctx->pid;
  e->uid = (__u32)bpf_get_current_uid_gid();
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sys_enter_openat — fires on every openat() syscall.
 * args[1] is the filename pointer (const char __user *).
 * We skip events from root/system daemons (uid < 1000) to reduce noise.
 */
SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat(struct sys_enter_ctx *ctx) {
  __u32 uid = (__u32)bpf_get_current_uid_gid();
  if (uid < 1000)
    return 0;

  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_OPEN;
  e->pid = bpf_get_current_pid_tgid() >> 32;
  e->uid = uid;
  bpf_get_current_comm(&e->comm, sizeof(e->comm));
  bpf_probe_read_user_str(&e->path, sizeof(e->path),
                          (const char *)ctx->args[1]);

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sched_process_exit, fires when any process exits.
 * pid and comm come directly from the tracepoint context.
 */
SEC("tracepoint/sched/sched_process_exit")
int trace_exit(struct sched_exit_ctx *ctx) {
  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_EXIT;
  e->pid = (__u32)ctx->pid;
  e->uid = (__u32)bpf_get_current_uid_gid();
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sys_enter_connect — fires on every connect() call.
 * args[1] is a struct sockaddr __user *. Decoding it (AF_INET vs AF_INET6,
 * extracting ip+port) requires reading the sa_family field first, then
 * branching. For v0.1 we record which process called connect; the Go
 * daemon can read /proc/<pid>/net/tcp for the actual peer if needed.
 */
SEC("tracepoint/syscalls/sys_enter_connect")
int trace_connect(struct sys_enter_ctx *ctx) {
  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_CONNECT;
  e->pid = bpf_get_current_pid_tgid() >> 32;
  e->uid = (__u32)bpf_get_current_uid_gid();
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  bpf_ringbuf_submit(e, 0);
  return 0;
}
