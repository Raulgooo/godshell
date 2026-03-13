#pragma once

/*
 * Primitive types for BPF programs — defined manually because linux/types.h
 * does not expose them cleanly when compiling with -target bpf outside of a
 * kernel build context. This avoids vmlinux.h as a build dependency.
 */

/* Unsigned */
typedef unsigned long long __u64;
typedef unsigned int __u32;
typedef unsigned short __u16;
typedef unsigned char __u8;

/* Signed */
typedef long long __s64;
typedef int __s32;
typedef short __s16;
typedef signed char __s8;

/* Big-endian network types (used by bpf_helper_defs.h networking helpers) */
typedef unsigned short __be16;
typedef unsigned int __be32;
typedef unsigned long long __be64;
typedef unsigned int __wsum;

#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <linux/bpf.h>

/* Networking types — defined here to avoid header conflicts and vmlinux
 * dependency */

#define AF_INET 2
#define AF_INET6 10

struct in_addr {
  __u32 s_addr;
};

struct sockaddr_in {
  __u16 sin_family;
  __be16 sin_port;
  struct in_addr sin_addr;
};

struct in6_addr {
  __u8 s6_addr[16];
};

struct sockaddr_in6 {
  __u16 sin6_family;
  __be16 sin6_port;
  __u32 sin6_flowinfo;
  struct in6_addr sin6_addr;
  __u32 sin6_scope_id;
};

struct sockaddr {
  __u16 sa_family;
  char sa_data[14];
};

#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohs(x) __builtin_bswap16(x)
#else
#define bpf_ntohs(x) (x)
#endif
