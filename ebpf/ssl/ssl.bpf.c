// go:build ignore

/*
 * ssl.bpf.c — Godshell SSL/TLS interceptor via uprobes
 *
 * Build (from project root):
 *   clang -g -O2 -target bpf -D__TARGET_ARCH_x86 \
 *     -I/usr/include/bpf -I./ebpf \
 *     -c ebpf/ssl/ssl.bpf.c -o ebpf/ssl/ssl_bpfel.o
 *
 * Attaches to:
 *   - libssl.so:  SSL_write / SSL_read
 *   - libnss3.so: PR_Write / PR_Read
 *   - Go binaries: crypto/tls.(*Conn).Write / crypto/tls.(*Conn).Read
 */

/* common.h defines all __u* types and includes linux/bpf.h + bpf_helpers.h */
#include "../../ebpf/common.h"

/* bpf_tracing.h provides SEC(), PT_REGS_PARM*, PT_REGS_RC */
#include <bpf/bpf_tracing.h>

/*
 * x86_64 pt_regs struct for uprobe programs.
 * bpf_tracing.h forward-declares 'struct pt_regs' but leaves it incomplete.
 * We provide the full definition here.
 */
#ifndef __ARCH_X86_64_PT_REGS_DEFINED
#define __ARCH_X86_64_PT_REGS_DEFINED
struct pt_regs {
  unsigned long r15, r14, r13, r12, rbp, rbx;
  unsigned long r11, r10, r9, r8;
  unsigned long rax, rcx, rdx, rsi, rdi;
  unsigned long orig_rax, rip, cs, eflags, rsp, ss;
};
#endif

char __license[] SEC("license") = "Dual MIT/GPL";

/* ── Constants ─────────────────────────────────────────────────────────── */
#define SSL_DIR_WRITE 0
#define SSL_DIR_READ 1
#define SSL_MAX_DATA 4096

/* ── SSL Event (kernel → userspace via ring buffer) ────────────────────── */
struct ssl_event {
  __u64 ts;
  __u32 pid;
  __u32 tid;
  __u8 direction;
  __u8 _pad[3];
  __u32 data_len;
  char comm[16];
  char data[SSL_MAX_DATA];
};

/* ── Maps ──────────────────────────────────────────────────────────────── */

struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 4 * 1024 * 1024);
} ssl_events SEC(".maps");

/* TID → buf pointer: stash at uprobe entry, consume at uretprobe exit */
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __type(key, __u32);
  __type(value, __u64);
  __uint(max_entries, 8192);
} ssl_buf_ptrs SEC(".maps");

/* ── Emit helper ───────────────────────────────────────────────────────── */
static __always_inline void emit_ssl(__u32 pid, __u32 tid, __u8 dir,
                                     const void *buf, __u32 len) {
  struct ssl_event *e = bpf_ringbuf_reserve(&ssl_events, sizeof(*e), 0);
  if (!e)
    return;
  /* Zero only the header fields — data[] gets overwritten by probe_read */
  e->ts = bpf_ktime_get_ns();
  e->pid = pid;
  e->tid = tid;
  e->direction = dir;
  // No _pad field to zero out explicitly after removal
  bpf_get_current_comm(&e->comm, sizeof(e->comm));
  __u32 cap = len < SSL_MAX_DATA ? len : SSL_MAX_DATA;
  e->data_len = cap;
  if (cap > 0)
    bpf_probe_read_user(e->data, cap, buf);
  bpf_ringbuf_submit(e, 0);
}

/* ── OpenSSL: SSL_write(SSL *ssl, const void *buf, int num) ────────────── */

SEC("uprobe/SSL_write")
int uprobe_ssl_write(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  emit_ssl(id >> 32, (__u32)id, SSL_DIR_WRITE, (void *)PT_REGS_PARM2(ctx),
           (__u32)PT_REGS_PARM3(ctx));
  return 0;
}

/* SSL_read: buf is filled AFTER the call; save ptr at entry, emit at exit */
SEC("uprobe/SSL_read")
int uprobe_ssl_read_entry(struct pt_regs *ctx) {
  __u32 tid = (__u32)bpf_get_current_pid_tgid();
  __u64 p = PT_REGS_PARM2(ctx);
  bpf_map_update_elem(&ssl_buf_ptrs, &tid, &p, BPF_ANY);
  return 0;
}

SEC("uretprobe/SSL_read")
int uretprobe_ssl_read(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  __u32 pid = id >> 32, tid = (__u32)id;
  __s32 ret = (__s32)PT_REGS_RC(ctx);
  if (ret <= 0) {
    bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
    return 0;
  }
  __u64 *bp = bpf_map_lookup_elem(&ssl_buf_ptrs, &tid);
  if (!bp)
    return 0;
  emit_ssl(pid, tid, SSL_DIR_READ, (void *)*bp, (__u32)ret);
  bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
  return 0;
}

/* ── NSS (Firefox): PR_Write(PRFileDesc *fd, const void *buf, PRInt32 amount)
 */

SEC("uprobe/PR_Write")
int uprobe_pr_write(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  emit_ssl(id >> 32, (__u32)id, SSL_DIR_WRITE, (void *)PT_REGS_PARM2(ctx),
           (__u32)PT_REGS_PARM3(ctx));
  return 0;
}

SEC("uprobe/PR_Read")
int uprobe_pr_read_entry(struct pt_regs *ctx) {
  __u32 tid = (__u32)bpf_get_current_pid_tgid();
  __u64 p = PT_REGS_PARM2(ctx);
  bpf_map_update_elem(&ssl_buf_ptrs, &tid, &p, BPF_ANY);
  return 0;
}

SEC("uretprobe/PR_Read")
int uretprobe_pr_read(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  __u32 pid = id >> 32, tid = (__u32)id;
  __s32 ret = (__s32)PT_REGS_RC(ctx);
  if (ret <= 0) {
    bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
    return 0;
  }
  __u64 *bp = bpf_map_lookup_elem(&ssl_buf_ptrs, &tid);
  if (!bp)
    return 0;
  emit_ssl(pid, tid, SSL_DIR_READ, (void *)*bp, (__u32)ret);
  bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
  return 0;
}

/*
 * Go crypto/tls — AMD64 register ABI (Go 1.17+)
 *   (*Conn).Write(p []byte): rdi=self, rsi=buf_ptr, rdx=len, rcx=cap
 *   (*Conn).Read(b []byte):  rdi=self, rsi=buf_ptr, rdx=len, rcx=cap
 *
 * Symbol: crypto/tls.(*Conn).Write and crypto/tls.(*Conn).Read
 * PT_REGS_PARM2 = rsi = buf_ptr
 * PT_REGS_PARM3 = rdx = len
 */

SEC("uprobe/go_tls_write")
int uprobe_go_tls_write(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  emit_ssl(id >> 32, (__u32)id, SSL_DIR_WRITE, (void *)PT_REGS_PARM2(ctx),
           (__u32)PT_REGS_PARM3(ctx));
  return 0;
}

SEC("uprobe/go_tls_read")
int uprobe_go_tls_read_entry(struct pt_regs *ctx) {
  __u32 tid = (__u32)bpf_get_current_pid_tgid();
  __u64 p = PT_REGS_PARM2(ctx);
  bpf_map_update_elem(&ssl_buf_ptrs, &tid, &p, BPF_ANY);
  return 0;
}

SEC("uretprobe/go_tls_read")
int uretprobe_go_tls_read(struct pt_regs *ctx) {
  __u64 id = bpf_get_current_pid_tgid();
  __u32 pid = id >> 32, tid = (__u32)id;
  __s32 ret = (__s32)PT_REGS_RC(ctx);
  if (ret <= 0) {
    bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
    return 0;
  }
  __u64 *bp = bpf_map_lookup_elem(&ssl_buf_ptrs, &tid);
  if (!bp)
    return 0;
  emit_ssl(pid, tid, SSL_DIR_READ, (void *)*bp, (__u32)ret);
  bpf_map_delete_elem(&ssl_buf_ptrs, &tid);
  return 0;
}
