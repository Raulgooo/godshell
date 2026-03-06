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
