// go:build ignore

#include "common.h"

char __license[] SEC("license") = "Dual MIT/GPL";

/* ── Constants ──────────────────────────────────────────────────────────── */

#define EVENT_EXEC 0
#define EVENT_OPEN 1
#define EVENT_EXIT 2
#define EVENT_CONNECT 3

/* ── Self-PID filter — set by Go before loading ─────────────────────────── *
 *
 * The Go daemon writes os.Getpid() into this variable via CollectionSpec
 * before loading the BPF objects. All handlers check it to drop self-events
 * in kernel space rather than userspace.
 */
volatile const __u32 godshell_pid = 0;

/* ── Shared event struct (kernel → userspace via ring buffer) ────────────── *
 *
 * Fixed-size. Go reads this byte-for-byte off the ring buffer and
 * deserializes it into the matching observer.Event struct.
 *
 * path is reused per event type:
 *   EVENT_EXEC    → binary path from execve args[0]
 *   EVENT_OPEN    → the file path passed to openat()
 *   EVENT_EXIT    → unused
 *   EVENT_CONNECT → unused
 */
struct event {
  __u64 ts;
  __u32 pid;
  __u32 uid;
  __u8 type;
  char comm[16];
  char path[256];
};

/* ── Maps ───────────────────────────────────────────────────────────────── */

/* Ring buffer for kernel → Go communication. */
struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 256 * 1024);
} events SEC(".maps");

/* ── Tracepoint context structs ──────────────────────────────────────────── */

/* sched/sched_process_exit/format */
struct sched_exit_ctx {
  unsigned short common_type;
  unsigned char common_flags;
  unsigned char common_preempt_count;
  int common_pid;
  char comm[16];
  int pid;
  int prio;
};

/* Generic sys_enter context — covers openat, connect, execve.
 * args[0..5] = syscall arguments. */
struct sys_enter_ctx {
  unsigned short common_type;
  unsigned char common_flags;
  unsigned char common_preempt_count;
  int common_pid;
  long __syscall_nr;
  unsigned long args[6];
};

/* ── Helpers ────────────────────────────────────────────────────────────── */

static __always_inline __u64 now(void) { return bpf_ktime_get_ns(); }

static __always_inline int is_self(__u32 pid) { return pid == godshell_pid; }

/* ── Handlers ───────────────────────────────────────────────────────────── */

/*
 * sys_enter_execve — fires BEFORE execve() runs.
 *
 * args[0] = filename  (const char __user *)
 * args[1] = argv      (const char __user * __user *)
 *
 * We capture the binary path from args[0] into e->path.
 * Full argv concatenation in kernel space fails the verifier due to
 * variable-offset buffer access; argv enrichment (if needed) is done
 * in the Go daemon via /proc/<pid>/cmdline after exec completes.
 */
SEC("tracepoint/syscalls/sys_enter_execve")
int trace_execve(struct sys_enter_ctx *ctx) {
  __u32 pid = bpf_get_current_pid_tgid() >> 32;
  if (is_self(pid))
    return 0;

  __u32 uid = (__u32)bpf_get_current_uid_gid();
  if (uid < 1000)
    return 0;

  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_EXEC;
  e->pid = pid;
  e->uid = uid;
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  /* Read the executable path (argv[0] equivalent) from the syscall arg */
  bpf_probe_read_user_str(e->path, sizeof(e->path), (const char *)ctx->args[0]);

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sys_enter_openat — fires on every openat() syscall.
 * args[1] is the filename pointer (const char __user *).
 */
SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat(struct sys_enter_ctx *ctx) {
  __u32 pid = bpf_get_current_pid_tgid() >> 32;
  if (is_self(pid))
    return 0;

  __u32 uid = (__u32)bpf_get_current_uid_gid();
  if (uid < 1000)
    return 0;

  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_OPEN;
  e->pid = pid;
  e->uid = uid;
  bpf_get_current_comm(&e->comm, sizeof(e->comm));
  bpf_probe_read_user_str(&e->path, sizeof(e->path),
                          (const char *)ctx->args[1]);

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sched_process_exit — fires when any process exits.
 */
SEC("tracepoint/sched/sched_process_exit")
int trace_exit(struct sched_exit_ctx *ctx) {
  __u32 pid = (__u32)ctx->pid;
  if (is_self(pid))
    return 0;

  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_EXIT;
  e->pid = pid;
  e->uid = (__u32)bpf_get_current_uid_gid();
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  bpf_ringbuf_submit(e, 0);
  return 0;
}

/*
 * sys_enter_connect — fires on every connect() call.
 */
SEC("tracepoint/syscalls/sys_enter_connect")
int trace_connect(struct sys_enter_ctx *ctx) {
  __u32 pid = bpf_get_current_pid_tgid() >> 32;
  if (is_self(pid))
    return 0;

  struct event *e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
  if (!e)
    return 0;

  __builtin_memset(e, 0, sizeof(*e));
  e->ts = now();
  e->type = EVENT_CONNECT;
  e->pid = pid;
  e->uid = (__u32)bpf_get_current_uid_gid();
  bpf_get_current_comm(&e->comm, sizeof(e->comm));

  bpf_ringbuf_submit(e, 0);
  return 0;
}
